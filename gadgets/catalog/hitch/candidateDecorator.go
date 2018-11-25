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
) (*api.ModuleCatalog, error) {
	// Load main catalog first.
	modCat, err := cat.ViewCatalogDelegate(ctx, modName)
	switch errcat.Category(err) {
	case nil:
		// continue!
	case hitch.ErrNoSuchCatalog:
		candidateModCat, err2 := cat.CandidateTree.LoadModuleCatalog(modName)
		switch errcat.Category(err2) {
		case nil:
			return candidateModCat, nil
		case hitch.ErrNoSuchCatalog:
			return nil, err // better to return the "not found" from the delegate.
		default:
			return nil, err2
		}
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
