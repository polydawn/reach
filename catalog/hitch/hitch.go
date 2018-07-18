// Package hitch provides an implementation backed by the filesystem for the
// interfaces specified by the go.polydawn.net/go-timeless-api/hitch package.
package hitch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/polydawn/go-errcat"
	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/stellar/catalog"
)

// FUTURE: a stateful delegating wrapper that injects "candidate" build data.

// FUTURE: a caching wrapper.

var (
	_ hitch.ViewCatalogTool    = FSCatalog{}.ViewCatalog
	_ hitch.ViewWarehousesTool = FSCatalog{}.ViewWarehouses
)

type FSCatalog struct {
	Root string
}

func (cat FSCatalog) ViewCatalog(
	ctx context.Context,
	modName api.ModuleName,
) (modCat *api.ModuleCatalog, err error) {
	return catalog.Tree{cat.Root}.LoadModuleCatalog(modName)
}

func (cat FSCatalog) ViewWarehouses(
	ctx context.Context,
	modName api.ModuleName,
) (ws *api.WareSourcing, err error) {
	funcs.MustValidate(modName)

	modPath := filepath.Join(cat.Root, string(modName))
	if fi, err := os.Stat(modPath); err != nil || !fi.IsDir() {
		return nil, errcat.Errorf(hitch.ErrNoSuchCatalog, "module not found: %q is not a dir", modName)
	}
	mirrorFilePath := filepath.Join(modPath, "mirrors.tl")
	if fi, err := os.Stat(mirrorFilePath); err != nil || !(fi.Mode()&os.ModeType == 0) {
		return &api.WareSourcing{}, nil
	}
	f, err := os.Open(mirrorFilePath)
	if err != nil {
		return nil, errcat.ErrorDetailed(hitch.ErrCorruptState, fmt.Sprintf("module mirrors list for %q cannot be opened: %s", api.ItemRef{modName, "", ""}.String(), err),
			map[string]string{
				"ref": api.ItemRef{modName, "", ""}.String(),
			})
	}
	defer f.Close()
	err = refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, f, api.Atlas_WareSourcing).Unmarshal(&ws)
	if err != nil {
		return nil, errcat.ErrorDetailed(hitch.ErrCorruptState, fmt.Sprintf("module mirrors list for %q cannot be opened: %s", api.ItemRef{modName, "", ""}.String(), err),
			map[string]string{
				"ref": api.ItemRef{modName, "", ""}.String(),
			})
	}
	return

}
