package layout

import (
	"fmt"
	"os"
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
)

// Landmarks holds all the major context-defining filesystem paths.
// Either module, or workspace, or both, or neither may be defined.
// (If it's neither, you're probably not getting much done, of course.)
//
// If FindLandmarks started with a relative path, all landmarks will be
// relative; if it was absolute, they'll all be absolute.
// The "*InsideWorkspace" paths are always relative and do not start with "./".
type Landmarks struct {
	ModuleRoot                string                // Path of module root (dir contains .timeless and (probably) module.tl), if any.
	ModuleFile                string                // Path of the module file (typically $moduleRoot/module.tl
	ModuleCatalogRoot         string                // Path to the module catalog root (typically $moduleRoot/.timeless/catalog/).
	WorkspaceRoot             string                // Path of the workspace root (dir contains .timeless and workspace.tl), if any.
	PathInsideWorkspace       string                // Path we are 'at' inside the workspaceRoot.  May be deeper than ModulePathInsideWorkspace; does not imply ModulePathInsideWorkspace is set.
	ModulePathInsideWorkspace string                // Path that contains 'module.tl' as relativized workspace.  Empty if no WorkspaceRoot.  Also empty if the module and the workspace root are the same path(!) (as this is definitely an anonymous module).  Is likely the module's name, but you should run that through workspace config.
	StagingWarehouse          api.WarehouseLocation // Address for a local warehouse (ca+file) where we'll store intermediates.
}

// FindLandmarks walks up the given path and looks for landmark files and dirs.
//
// The path can be relative or absolute; results will be in the same format
// (but always clean'd).  If the given path is relative, the search will not
// recurse above it.
//
// If there is more than one module found, the closest (i.e. longest path) one
// will be reported.
// If a workspace is found before a module is, FindLandmarks terminates search
// and returns that workspace path alone.
// In the edge case that a workspace and a module are coresident, both will
// be detected correctly.
func FindLandmarks(startPath string) (*Landmarks, error) {
	startClean := filepath.Clean(startPath)
	// Set fallback defaults before starting.
	marks := &Landmarks{
		StagingWarehouse: "ca+file://./.timeless/warehouse/",
	}
	// Walk up the path, noting any landmarks as we go,
	//  terminating when we run out of path segments.
	dir := startClean
	for {
		// `ls`.  Any errors: return.
		f, err := os.Open(dir)
		if err != nil {
			return marks, err
		}
		// Scan through all entries in the dir, looking for our landmarks.
		names, err := f.Readdirnames(-1)
		f.Close()
		if err != nil {
			return marks, err
		}
		dirHasDotTimeless := false
		dirHasKnownRole := false
		for _, name := range names {
			switch name {
			case "module.tl", "workspace.tl", ".timeless":
				// interesting!
			default:
				continue
			}
			pth := filepath.Join(dir, name)
			fi, err := os.Stat(pth)
			if err != nil {
				return marks, err
			}
			switch name {
			case "module.tl":
				if fi.Mode()&os.ModeType != 0 {
					return marks, fmt.Errorf("'module.tl' should be a file (%q)", pth)
				}
				dirHasKnownRole = true
				if marks.ModuleRoot == "" {
					marks.ModuleRoot = dir
					marks.ModuleFile = pth
					marks.ModuleCatalogRoot = filepath.Join(dir, ".timeless/catalog")
				}
			case "workspace.tl":
				if fi.Mode()&os.ModeType != 0 {
					return marks, fmt.Errorf("'workspace.tl' should be a file (%q)", pth)
				}
				dirHasKnownRole = true
				marks.WorkspaceRoot = dir
				marks.PathInsideWorkspace = startClean[len(dir):]
				marks.StagingWarehouse = api.WarehouseLocation("ca+file://" + dir + "/.timeless/warehouse")
			case ".timeless":
				if !fi.IsDir() {
					return marks, fmt.Errorf("'.timeless' should be a dir (%q)", pth)
				}
				dirHasDotTimeless = true
			default:
				panic("unreachable")
			}
		}
		// If we found a '.timeless' dir, but no other purpose-landmarking file,
		//  check for special cases in deeper paths... and if that doesn't yield,
		//   then we're looking at a layout that seems to be speaking to us,
		//    but we don't know why so we should probably halt and report.
		if dirHasDotTimeless && !dirHasKnownRole {
			pth := filepath.Join(dir, ".timeless/module.tl")
			fi, err := os.Stat(pth)
			if os.IsNotExist(err) {
				return marks, fmt.Errorf("'.timeless' dir found but unaccompanied at %q; maybe you want to make a module.tl or workspace.tl file?", dir)
			}
			if err != nil {
				return marks, err
			}
			if fi.Mode()&os.ModeType != 0 {
				return marks, fmt.Errorf("'module.tl' should be a file (%q)", pth)
			}
			if marks.ModuleRoot == "" {
				marks.ModuleRoot = dir
				marks.ModuleFile = pth
				marks.ModuleCatalogRoot = filepath.Join(dir, ".timeless/catalog")
			}
		}
		// If we found a workspace, woot, let's go home.
		if marks.WorkspaceRoot != "" {
			// But first, derive any more {foo}-relative-to-workspace paths.
			if marks.ModuleRoot != "" {
				if len(marks.ModuleRoot) == len(marks.WorkspaceRoot) {
					marks.ModulePathInsideWorkspace = ""
				} else {
					marks.ModulePathInsideWorkspace = marks.ModuleRoot[len(marks.WorkspaceRoot)+1:]
				}
			}
			return marks, nil
		}
		// If basename'ing got us "/" this time, and we still didn't find it, terminate.
		if dir == "/" || dir == "." {
			return marks, nil
		}
		// Step up one dir.
		dir = filepath.Dir(dir)
	}
}
