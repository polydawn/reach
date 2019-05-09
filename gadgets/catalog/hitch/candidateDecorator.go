package hitch

import (
	"context"

	"github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/reach/gadgets/catalog"
)

func WithCandidates(viewTool hitch.ViewLineageTool, candidatesTree catalog.Tree) hitch.ViewLineageTool {
	return candidateDecorator{viewTool, candidatesTree}.ViewLineage
}

type candidateDecorator struct {
	ViewLineageDelegate hitch.ViewLineageTool
	CandidateTree       catalog.Tree // Releases here are prepended to others.
}

func (cat candidateDecorator) ViewLineage(
	ctx context.Context,
	modName api.ModuleName,
) (*api.Lineage, error) {
	// Load main catalog first.
	lin, err := cat.ViewLineageDelegate(ctx, modName)
	switch errcat.Category(err) {
	case nil:
		// continue!
	case hitch.ErrNoSuchLineage:
		candidateLin, err2 := cat.CandidateTree.LoadModuleLineage(modName)
		switch errcat.Category(err2) {
		case nil:
			return candidateLin, nil
		case hitch.ErrNoSuchLineage:
			return nil, err // better to return the "not found" from the delegate.
		default:
			return nil, err2
		}
	default:
		return nil, err
	}
	// Load any candidate info and if it exists merge it in.
	candidateLin, err := cat.CandidateTree.LoadModuleLineage(modName)
	switch errcat.Category(err) {
	case nil:
		// continue!
	case hitch.ErrNoSuchLineage:
		return lin, nil
	default:
		return nil, err
	}
	rel, err := hitch.LineagePluckReleaseByName(*candidateLin, "candidate")
	if err != nil {
		panic("a lineage in the candidate tree may only contain a release called \"candidate\"")
	}
	lin, err = hitch.LineagePrependRelease(*lin, *rel)
	if err != nil {
		panic("a lineage in a catalog shouldn't already contain a \"candidate\"!!")
	}
	return lin, nil
}
