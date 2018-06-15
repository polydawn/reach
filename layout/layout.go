package layout

import (
	"fmt"
	"os"
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
)

type TreeInfo struct {
	Root        string
	Singleton   bool
	CatalogRoot string
}

func FindTree(startPath string) (*TreeInfo, error) {
	// Look for root.
	root, found, err := getDirContainingMarkerDir(startPath, ".timeless")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no .timeless dir found")
	}
	root += "/.timeless"
	// Figure out if it's a singleton module or a whole stellar neighborhood.
	fi, err := os.Stat(filepath.Join(root, "module.tl"))
	if err == nil {
		if fi.Mode()&os.ModeType != 0 {
			return nil, fmt.Errorf("%s has no known layout (if 'module.tl' exists, should be file)", root)
		}
		return &TreeInfo{
			Root:        root,
			Singleton:   true,
			CatalogRoot: filepath.Join(root, "catalog"),
		}, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	// FUTURE consider other layouts.
	return nil, fmt.Errorf("%s has no known layout (expecting a 'module.tl' file)", root)
}

func LoadModule(ti TreeInfo, modName api.ModuleName) (*api.Module, error) {
	panic("TODO")
}

/*
	Look for a workspace indicated by a directory with special name:
	starting from `startPath`,
	iterating up through parent directories,
	looking for dirs containing another dir named `marker`,
	and returning the first such dir found; or, any errors encountered,
	or false if we reached the root or have no more startPath to check.
*/
func getDirContainingMarkerDir(startPath string, marker string) (dirFound string, found bool, err error) {
	dir := filepath.Clean(startPath)
	for {
		// `ls`.  Any errors: return.
		f, err := os.Open(dir)
		if err != nil {
			return "", false, err
		}
		// Scan through all entries in the dir, looking for our fav.
		names, err := f.Readdirnames(-1)
		f.Close()
		if err != nil {
			return "", false, err
		}
		for _, name := range names {
			if name != marker {
				continue
			}
			pth := filepath.Join(dir, name)
			fi, err := os.Stat(pth)
			if err != nil {
				return "", false, err
			}
			if !fi.IsDir() {
				break
			}
			return dir, true, nil
		}
		// If basename'ing got us "/" this time, and we still didn't find it, terminate.
		if dir == "/" || dir == "." {
			return "", false, nil
		}
		// Step up one dir.
		dir = filepath.Dir(dir)
	}
}
