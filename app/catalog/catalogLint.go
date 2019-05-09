package catalogApp

import (
	"fmt"
	"os"
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/reach/gadgets/catalog"
)

type Linter struct {
	Tree         catalog.Tree
	WarnBehavior func(msg string, remedy func())
	Rewrite      bool
}

func (cfg Linter) Lint() error {
	err := filepath.Walk(cfg.Tree.Root, func(path string, info os.FileInfo, err error) error {
		// Root dir is a special case.  Check it's at least a dir, then skip.
		if path == cfg.Tree.Root {
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeType != os.ModeDir {
				return fmt.Errorf("catalog lint: did not start at directory (%s)", path)
			}
			return nil
		}
		// Get the path sans the prefix of the root.
		modulePath := path[len(cfg.Tree.Root)+1:]
		// Ignore dotfiles at the root.  (.git is not unlikely here)
		if modulePath[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// And now on with your regularly scheduled programming.
		switch info.Mode() & ^os.ModePerm {
		case 0: // file
			basename := filepath.Base(path)
			moduleName := api.ModuleName(modulePath[:len(modulePath)-len(basename)-1])
			switch basename {
			case catalog.MirrorsFileName:
				// Check parse.
				ws, err := cfg.Tree.LoadModuleMirrors(moduleName)
				if err != nil {
					cfg.WarnBehavior(fmt.Sprintf("%v", err), func() {})
					return nil
				}
				if ws == nil {
					return nil
				}

				// Check semantic sanity.
				//  Uses a foolish probably-duplicate load of catalog.
				lin, err := cfg.Tree.LoadModuleLineage(moduleName)
				if err != nil {
					return nil // skip the rest of this check and wait for that error to be rediscovered later.
				}
				// Check that all info refers to our own module, and none is dangling nor overreaching.
				if len(ws.ByPackType) > 0 {
					cfg.WarnBehavior(
						fmt.Sprintf("in mirror list for %q, 'ByPackType' should not be used; use 'ByModule' instead", moduleName),
						func() {
							for pktype, whs := range ws.ByPackType {
								ws.AppendByModule(moduleName, pktype, whs...)
							}
							ws.ByPackType = nil
						},
					)
				}
				for modName, rest := range ws.ByModule {
					if modName != moduleName {
						cfg.WarnBehavior(
							fmt.Sprintf("in mirror list for %q, 'ByModule' refers to another module, which is silly", moduleName),
							func() {
								for pktype, whs := range rest {
									ws.AppendByModule(moduleName, pktype, whs...)
								}
								delete(ws.ByModule, modName)
							},
						)
					}
				}
				allCatalogWares := map[api.WareID]struct{}{}
				for _, rel := range lin.Releases {
					for _, wareID := range rel.Items {
						allCatalogWares[wareID] = struct{}{}
					}
				}
				for wareID, _ := range ws.ByWare {
					if _, present := allCatalogWares[wareID]; !present {
						cfg.WarnBehavior(
							fmt.Sprintf("in mirror list for %q, 'ByWare' refers to a ware %q which is not actually in the catalog", moduleName, wareID),
							func() {
								delete(ws.ByWare, wareID)
							},
						)
					}
				}
				// FUTURE we don't yet lint for wares in a catalog but have no suggestions of any warehouses.
				//  We could (although also it would be sort of a partial defense, because we're not going to check actual availability from here).

				// Rewrite, ensuring bytewise normality.
				if cfg.Rewrite {
					cfg.Tree.SaveModuleMirrors(moduleName, *ws)
				}
			case catalog.LineageFileName:
				// Check parse.
				lin, err := cfg.Tree.LoadModuleLineage(moduleName)
				if err != nil {
					cfg.WarnBehavior(fmt.Sprintf("%v", err), func() {})
					return nil
				}

				// Check semantic sanity.
				// Check that the file's concept of who it is matches the path.
				if lin.Name != moduleName {
					cfg.WarnBehavior(
						fmt.Sprintf("in lineage for module %q, moduleName does not match path!", moduleName),
						func() {
							lin.Name = moduleName
						},
					)
				}
				// Check that all release names are valid, and no dupliline entries.
				takenNames := map[api.ReleaseName]struct{}{}
				for _, rel := range lin.Releases {
					// TODO validation rule for release names missing
					if _, present := takenNames[rel.Name]; present {
						cfg.WarnBehavior(
							fmt.Sprintf("in lineage for module %q, multiple releases found named %q!", moduleName, rel.Name),
							func() {
								lin.Name = moduleName
							},
						)
					}
					takenNames[rel.Name] = struct{}{}
				}

				// Rewrite, ensuring bytewise normality.
				if cfg.Rewrite {
					cfg.Tree.SaveModuleLineage(moduleName, *lin)
				}
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
	// TODO some linegorization of any walk errors.
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
