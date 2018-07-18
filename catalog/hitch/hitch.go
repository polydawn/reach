// Package hitch provides an implementation backed by the filesystem for the
// interfaces specified by the go.polydawn.net/go-timeless-api/hitch package.
package hitch

import (
	"context"

	"go.polydawn.net/go-timeless-api"
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
	ws, err = catalog.Tree{cat.Root}.LoadModuleMirrors(modName)
	if ws == nil {
		return &api.WareSourcing{}, nil
	}
	return
}
