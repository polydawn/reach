package hitch

import (
	"context"

	"github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/stellar/gadgets/catalog"
)

func WithCandidates(viewTool hitch.ViewCatalogTool, candidatesTree catalog.Tree) hitch.ViewCatalogTool {
	return candidateDecorator{viewTool, candidatesTree}.ViewCatalog
}

type candidateDecorator struct {
	ViewCatalogDelegate hitch.ViewCatalogTool
	CandidateTree       catalog.Tree // Releases here are prepended to others.
}

func (cat candidateDecorator) ViewCatalog(
	ctx context.Context,
	modName api.ModuleName,
) (modCat *api.ModuleCatalog, err error) {
	modCat, err = cat.ViewCatalogDelegate(ctx, modName)
	switch errcat.Category(err) {
	case nil:
		return modCat, nil
	case hitch.ErrNoSuchCatalog:
		return cat.CandidateTree.LoadModuleCatalog(modName)
	default:
		return nil, err
	}
}
