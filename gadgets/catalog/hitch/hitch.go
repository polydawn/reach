// Package hitch provides an implementation backed by the filesystem for the
// interfaces specified by the go.polydawn.net/go-timeless-api/hitch package.
package hitch

import (
	"context"

	"github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/reach/gadgets/catalog"
)

// FUTURE: a caching wrapper.

func ViewTools(trees ...catalog.Tree) (
	hitch.ViewCatalogTool,
	hitch.ViewWarehousesTool,
) {
	cat := basicCatalog{trees}
	return cat.ViewCatalog, cat.ViewWarehouses
}

type basicCatalog struct {
	Trees []catalog.Tree // Reads probe linearly down.
}

func (cat basicCatalog) ViewCatalog(
	ctx context.Context,
	modName api.ModuleName,
) (modCat *api.ModuleCatalog, err error) {
	for _, tree := range cat.Trees {
		modCat, err = tree.LoadModuleCatalog(modName)
		switch errcat.Category(err) {
		case nil:
			return modCat, nil
		case hitch.ErrNoSuchCatalog:
			continue
		default:
			return nil, err
		}
	}
	return
}

func (cat basicCatalog) ViewWarehouses(
	ctx context.Context,
	modName api.ModuleName,
) (ws *api.WareSourcing, err error) {
	for _, tree := range cat.Trees {
		ws, err = tree.LoadModuleMirrors(modName)
		switch errcat.Category(err) {
		case nil:
			if ws == nil {
				return &api.WareSourcing{}, nil
			}
			return ws, nil
		case hitch.ErrNoSuchCatalog:
			continue
		default:
			return nil, err
		}
	}
	return
}
