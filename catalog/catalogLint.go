package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
)

type Linter struct {
	Tree         Tree
	WarnBehavior func(msg string, remedy func())
}

func (cfg Linter) Lint() error {
	err := filepath.Walk(cfg.Tree.Root, func(path string, info os.FileInfo, err error) error {
		modulePath := path[len(cfg.Tree.Root):]
		switch info.Mode() & ^os.ModePerm {
		case 0: // file
			basename := filepath.Base(path)
			switch basename {
			case "mirrors.tl":
				// Check parse.
				// TODO

				// Check semantic sanity.
				// TODO
				// FUTURE doing full sanity checks of the data itself rather than just the format
				//  will involve opening the catalog.tl file to check values against.

				// Rewrite, ensuring bytewise normality.
				// TODO
			case "catalog.tl":
				// Check parse.
				// TODO

				// Check semantic sanity.
				// TODO

				// Rewrite, ensuring bytewise normality.
				// TODO
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
	// TODO some categorization of any walk errors.
	return err
}

// remove is a factory that binds the eponymous often-suggested remedy for
// linter warning about unexpected files, invalid path names, etc, to a
// specific path.
func remove(path string) func() {
	return func() {
		os.Remove(path)
	}
}
