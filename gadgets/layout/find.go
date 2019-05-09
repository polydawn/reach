package layout

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/warpfork/go-errcat"
)

type ErrorCategory string

const (
	WorkspaceSearchError = ErrorCategory("reach-workspace-search-error")
	WorkspaceNotFound    = ErrorCategory("reach-workspace-not-found")
	ModuleSearchError    = ErrorCategory("reach-module-search-error")
	ModuleNotFound       = ErrorCategory("reach-module-not-found")
)

/*
	For context: typical usage of these functions varies
	according to the CLI invocation style:

	- `reach emerge` --> start search for workspace and module at $PWD.

	- `reach emerge foo/bar` --> start search for workspace at $PWD;
	start search for module at '$PWD/foo/bar'.
	Note how this usage may find a different workspace than if you
	had run '(cd foo/bar && reach emerge)'!

	Potential future work: we might want to find *all* workspaces between
	$PWD and '$PWD/foo/bar', so that we can run consistency sanity checks.
	But this would be a feature which is not yet fully specified.
*/

// FindWorkspace looks for the '.timeless' dir indicating a workspace
// and returns a description of the found paths.
//
// The startPath and its parents will be searched until a workspace is found.
//
// The path can be relative or absolute; results will be in the same format
// (but always clean'd).  If the given path is relative, the search will not
// recurse above it.
func FindWorkspace(startPath string) (*Workspace, error) {
	startClean := filepath.Clean(startPath)
	for dir := startClean; ; dir = filepath.Dir(dir) {
		// `ls`.  Any errors: return.
		f, err := os.Open(dir)
		if err != nil {
			return nil, errcat.Errorf(WorkspaceSearchError, "%s", err)
		}
		names, err := f.Readdirnames(-1)
		f.Close()
		if err != nil {
			return nil, errcat.Errorf(WorkspaceSearchError, "%s", err)
		}

		// Scan through all entries in the dir, looking for our landmarks.
		for _, name := range names {
			if name != ".timeless" {
				continue
			}
			pth := filepath.Join(dir, name)
			fi, err := os.Stat(pth)
			if err != nil {
				return nil, errcat.Errorf(WorkspaceSearchError, "%s", err)
			}
			if !fi.IsDir() {
				return nil, errcat.Errorf(WorkspaceSearchError, "'.timeless' found (%q) but must be a dir", pth)
			}
			return &Workspace{
				dir,
			}, nil
		}

		// If basename'ing got us "/" this time, and we still didn't find it, terminate.
		if dir == "/" || dir == "." {
			return nil, errcat.Errorf(WorkspaceNotFound, "workspace not found")
		}
	}
}

// FindModule looks for the 'module.tl' file indicating a module
// and returns a description of the found paths.
//
// The startPath and its parents will be searched until a module is found.
//
// The path can be relative or absolute; results will be in the same format
// (but always clean'd).  If the given path is relative, the search will not
// recurse above it.
//
// If the startPath is not included under the workspace's root path, an error
// will be raised immediately -- it is not valid to have a module without
// a containing workspace.
func FindModule(lm Workspace, startPath string) (*Module, error) {
	startClean := filepath.Clean(startPath)

	if !strings.HasPrefix(startClean+"/", lm.workspaceRoot+"/") {
		return nil, errcat.Errorf(ModuleSearchError, "module path must be contained within a workspace root (workspace root is %q)", lm.workspaceRoot)
	}

	for dir := startClean; ; dir = filepath.Dir(dir) {
		// `ls`.  Any errors: return.
		f, err := os.Open(dir)
		if err != nil {
			return nil, errcat.Errorf(ModuleSearchError, "%s", err)
		}
		names, err := f.Readdirnames(-1)
		f.Close()
		if err != nil {
			return nil, errcat.Errorf(ModuleSearchError, "%s", err)
		}

		// Scan through all entries in the dir, looking for our landmarks.
		for _, name := range names {
			if name != "module.tl" {
				continue
			}
			pth := filepath.Join(dir, name)
			fi, err := os.Stat(pth)
			if err != nil {
				return nil, errcat.Errorf(ModuleSearchError, "%s", err)
			}
			if fi.Mode()&os.ModeType != 0 {
				return nil, errcat.Errorf(ModuleSearchError, "'module.tl' found (%q) but must be a file", pth)
			}
			return &Module{
				lm,
				dir,
			}, nil
		}

		// If basename'ing got us "/" this time, and we still didn't find it, terminate.
		if dir == "/" || dir == "." {
			return nil, errcat.Errorf(ModuleNotFound, "no module found")
		}
	}
}

// ExpectModule is similar to FindModule, but doesn't search; if there's
// no module in the path given, it returns an error immediately.
//
// (This is the behavior when using an explicit arg on the CLI rather than
// providing zero args and letting it find a module based on PWD.)
func ExpectModule(lm Workspace, pth string) (*Module, error) {
	dir := filepath.Clean(pth)

	if !strings.HasPrefix(dir+"/", lm.workspaceRoot+"/") {
		return nil, errcat.Errorf(ModuleSearchError, "module path must be contained within a workspace root (workspace root is %q)", lm.workspaceRoot)
	}

	// `ls`.  Any errors: return.
	f, err := os.Open(dir)
	if err != nil {
		return nil, errcat.Errorf(ModuleSearchError, "%s", err)
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, errcat.Errorf(ModuleSearchError, "%s", err)
	}

	// Scan through all entries in the dir, looking for our landmarks.
	for _, name := range names {
		if name != "module.tl" {
			continue
		}
		pth := filepath.Join(dir, name)
		fi, err := os.Stat(pth)
		if err != nil {
			return nil, errcat.Errorf(ModuleSearchError, "%s", err)
		}
		if fi.Mode()&os.ModeType != 0 {
			return nil, errcat.Errorf(ModuleSearchError, "'module.tl' found (%q) but must be a file", pth)
		}
		return &Module{
			lm,
			dir,
		}, nil
	}

	// If it wasn't in this dir, game over.  No search in this mode.
	return nil, errcat.Errorf(ModuleNotFound, "no module found")
}

// Future: `FindModulesBeneath(startPath) ([]Module)` -- we would use this to
//  handle commands like `reach emerge ./group/...`.
