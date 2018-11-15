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
	// Load main catalog first.
	modCat, err = cat.ViewCatalogDelegate(ctx, modName)
	switch errcat.Category(err) {
	case nil:
		// continue!
	case hitch.ErrNoSuchCatalog:
		return cat.CandidateTree.LoadModuleCatalog(modName)
	default:
		return nil, err
	}
	// Load any candidate info and if it exists merge it in.
	candidateModCat, err := cat.CandidateTree.LoadModuleCatalog(modName)
	switch errcat.Category(err) {
	case nil:
		// continue!
	case hitch.ErrNoSuchCatalog:
		return modCat, nil
	default:
		return nil, err
	}
	rel, err := hitch.CatalogPluckReleaseByName(*candidateModCat, "candidate")
	if err != nil {
		panic("a catalog in the candidate tree must only contain a release called \"candidate\"")
	}
	modCat, err = hitch.CatalogPrependRelease(*modCat, *rel)
	if err != nil {
		panic("main track release catalogs shouldn't already contain a \"candidate\"!")
	}
	return modCat, nil
}
