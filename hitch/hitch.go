// Package hitch provides an implementation backed by the filesystem for the
// interfaces specified by the go.polydawn.net/go-timeless-api/hitch package.
package hitch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/hitch"
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
	funcs.MustValidate(modName)

	modPath := filepath.Join(cat.Root, string(modName))
	if fi, err := os.Stat(modPath); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("module not found: %q is not a dir.", modName)
	}
	modCatPath := filepath.Join(modPath, "catalog.tl")
	if fi, err := os.Stat(modCatPath); err != nil || !(fi.Mode()&os.ModeType == 0) {
		return nil, fmt.Errorf("module catalog not found: no %s file under %q.", "catalog.tl", modName)
	}
	f, err := os.Open(modCatPath)
	if err != nil {
		return
	}
	defer f.Close()
	err = refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, f, api.Atlas_Catalog).Unmarshal(&modCat)
	return
}

func (cat FSCatalog) ViewWarehouses(
	ctx context.Context,
	modName api.ModuleName,
) (ws *api.WareSourcing, err error) {
	funcs.MustValidate(modName)

	modPath := filepath.Join(cat.Root, string(modName))
	if fi, err := os.Stat(modPath); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("module not found: %q is not a dir.", modName)
	}
	mirrorFilePath := filepath.Join(modPath, "mirrors.tl")
	if fi, err := os.Stat(mirrorFilePath); err != nil || !(fi.Mode()&os.ModeType == 0) {
		return &api.WareSourcing{}, nil
	}
	f, err := os.Open(mirrorFilePath)
	if err != nil {
		return
	}
	defer f.Close()
	err = refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, f, api.Atlas_WareSourcing).Unmarshal(&ws)
	return

}
