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
		if path == cfg.Tree.Root { // skip root dir
			return nil
		}
		modulePath := path[len(cfg.Tree.Root)+1:]
		if modulePath[0] == '.' { // ignore dotfiles at the root (.git is not unlikely here)
			return filepath.SkipDir
		}
		switch info.Mode() & ^os.ModePerm {
		case 0: // file
			basename := filepath.Base(path)
			moduleName := api.ModuleName(modulePath[:len(modulePath)-len(basename)-1])
			switch basename {
			case "mirrors.tl":
				// Check parse.
				ws, err := cfg.Tree.LoadModuleMirrors(moduleName)
				if err != nil {
					cfg.WarnBehavior(fmt.Sprintf("%v", err), func() {})
					return nil
				}

				// Check semantic sanity.
				// TODO
				_ = ws
				// FUTURE doing full sanity checks of the data itself rather than just the format
				//  will involve opening the catalog.tl file to check values against.

				// Rewrite, ensuring bytewise normality.
				// TODO
			case "catalog.tl":
				// Check parse.
				cat, err := cfg.Tree.LoadModuleCatalog(moduleName)
				if err != nil {
					cfg.WarnBehavior(fmt.Sprintf("%v", err), func() {})
					return nil
				}

				// Check semantic sanity.
				// TODO
				_ = cat

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
				cfg.WarnBehavior(fmt.Sprintf("dir %q is not a valid moduleName: %v", moduleName, err), remove(path))
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
