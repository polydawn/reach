package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/polydawn/go-errcat"
	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"github.com/polydawn/refmt/obj/atlas"
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
)

type Tree struct {
	Root string
}

// LoadModuleMirrors attempts to load the catalog.tl file for a module.
// The result can never be nil unless there is an error since the catalog file is required.
func (tree Tree) LoadModuleCatalog(modName api.ModuleName) (modCat *api.ModuleCatalog, err error) {
	err = tree.loadModuleFile(modName, &modCat, api.Atlas_Catalog, true, "catalog", "catalog.tl")
	return
}

// SaveModuleMirrors writes out a catalog.tl file.
// The dirs will be created if necessary.
func (tree Tree) SaveModuleCatalog(modName api.ModuleName, modCat api.ModuleCatalog) error {
	if err := modName.Validate(); err != nil {
		return errcat.ErrorDetailed(
			hitch.ErrUsage,
			fmt.Sprintf("cannot save catalog: %q is not a valid module name: %s", modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	if err := os.MkdirAll(filepath.Join(tree.Root, string(modName)), 0755); err != nil {
		return errcat.ErrorDetailed(
			hitch.ErrCorruptState,
			fmt.Sprintf("cannot save catalog for module %q: %s", modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	return tree.saveModuleFile(modName, modCat, api.Atlas_Catalog, "catalog", "catalog.tl")
}

// LoadModuleMirrors attempts to load the mirrors.tl file for a module.
// The result is nil and nil error iff the file does not exist.
func (tree Tree) LoadModuleMirrors(modName api.ModuleName) (ws *api.WareSourcing, err error) {
	err = tree.loadModuleFile(modName, &ws, api.Atlas_WareSourcing, false, "mirrors list", "mirrors.tl")
	return
}

// SaveModuleMirrors writes out a mirrors.tl file.
// The catalog must be written first (e.g. the dir must exist).
func (tree Tree) SaveModuleMirrors(modName api.ModuleName, ws api.WareSourcing) error {
	return tree.saveModuleFile(modName, ws, api.Atlas_WareSourcing, "mirrors list", "mirrors.tl")
}

//
// above: Load and Save methods for known files.
// --------
// below: fiddly helper bits.
//

// loadModuleFile does expectModuleFile, then attempts to load the content.
// This method currently presumes json format in the files.
func (tree Tree) loadModuleFile(modName api.ModuleName, structure interface{}, atl atlas.Atlas, required bool, purpose, filename string) error {
	f, err := tree.expectModuleFile(modName, required, purpose, filename)
	if err != nil {
		return err
	}
	if f == nil {
		return nil
	}
	defer f.Close()
	err = refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, f, atl).Unmarshal(structure)
	if err != nil {
		return errcat.ErrorDetailed(
			hitch.ErrCorruptState,
			fmt.Sprintf("module %s failed to parse for %q: %s", purpose, modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	return nil
}

// expectModuleFile does expectModule, then expects a particular file to exist,
// and returns a polite error with a message including purpose info if it fails.
func (tree Tree) expectModuleFile(modName api.ModuleName, required bool, purpose, filename string) (*os.File, error) {
	if err := tree.expectModule(modName); err != nil {
		return nil, err
	}
	modFilePath := filepath.Join(tree.Root, string(modName), filename)
	if fi, err := os.Stat(modFilePath); err != nil || !(fi.Mode()&os.ModeType == 0) {
		if required {
			return nil, errcat.ErrorDetailed(
				hitch.ErrCorruptState,
				fmt.Sprintf("module %s failed to load for %q: no %s file in module dir", purpose, modName, filename),
				map[string]string{
					"ref": string(modName),
				})
		}
		return nil, nil
	}
	f, err := os.Open(modFilePath)
	if err != nil {
		return nil, errcat.ErrorDetailed(
			hitch.ErrCorruptState,
			fmt.Sprintf("module %s failed to load for %q: cannot open %s file in module dir: %s", purpose, modName, filename, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	return f, nil
}

// expectModule returns a polite error if a module is not found in this tree.
// Use it before trying to open any particular files so we get a better message
// for absenses.
func (tree Tree) expectModule(modName api.ModuleName) error {
	if err := modName.Validate(); err != nil {
		return errcat.ErrorDetailed(
			hitch.ErrUsage,
			fmt.Sprintf("module %q: not a valid module name: %s", modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	modPath := filepath.Join(tree.Root, string(modName))
	if fi, err := os.Stat(modPath); err != nil || !fi.IsDir() {
		return errcat.ErrorDetailed(
			hitch.ErrNoSuchCatalog,
			fmt.Sprintf("module %q not found: %q is not a dir", modName, modPath),
			map[string]string{
				"ref": string(modName),
			})
	}
	return nil
}

// saveModuleFile does expectModule, then serializes and writes to the file.
// This method currently presumes json format in the files.
func (tree Tree) saveModuleFile(modName api.ModuleName, structure interface{}, atl atlas.Atlas, purpose, filename string) error {
	err := tree.expectModule(modName)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(tree.Root, string(modName), filename), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errcat.ErrorDetailed(
			hitch.ErrCorruptState,
			fmt.Sprintf("module %s failed to save for %q: %s", purpose, modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	defer f.Close()
	err = refmt.NewMarshallerAtlased(json.EncodeOptions{Line: []byte{'\n'}, Indent: []byte{'\t'}}, f, atl).Marshal(structure)
	if err != nil {
		return errcat.ErrorDetailed(
			hitch.ErrCorruptState, // This is actually kind of catastrophic and hopefully isn't reachable.
			fmt.Sprintf("module %s failed to save for %q: %s", purpose, modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	return nil
}
