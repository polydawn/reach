package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
)

type Linter struct {
	Root         string
	WarnBehavior func(msg string, remedy func())
}

func (cfg *Linter) Lint() {
	err := filepath.Walk(cfg.Root, func(path string, info os.FileInfo, err error) error {
		modulePath := path[len(cfg.Root):]
		switch info.Mode() & ^os.ModePerm {
		case 0: // file
			basename := filepath.Base(path)
			switch basename {
			case "mirrors.tl":
				// TODO read, sanity check, rewrite
				// FUTURE doing full sanity checks of the data itself rather than just the format
				//  will involve opening the catalog.tl file to check values against.
			case "catalog.tl":
				// TODO read, sanity check, rewrite
			default:
				// TODO warn about any files of names we don't know about
				// FUTURE need cases for matching the replay prefix
			}
			return nil
		case os.ModeDir:
			moduleName := api.ModuleName(modulePath)
			if err := moduleName.Validate(); err != nil {
				cfg.WarnBehavior(fmt.Sprintf("dir is not a valid moduleName: %v", err), remove(path))
				return filepath.SkipDir
			}
			// FUTURE empty dirs should get a warning
			return nil
		default:
			cfg.WarnBehavior(fmt.Sprintf("unexpected file of type %v", info.Mode()), remove(path))
			return nil
		}
	})
	_ = err
}

// remove is a factory that binds the eponymous often-suggested remedy for
// linter warning about unexpected files, invalid path names, etc, to a
// specific path.
func remove(path string) func() {
	return func() {
		os.Remove(path)
	}
}
