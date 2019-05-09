package catalog

import (
	"io"
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/reach/gadgets/layout"
)

func SaveCandidateRelease(landmarks layout.Workspace, sagaName SagaName, modName api.ModuleName, content map[api.ItemName]api.WareID, stderr io.Writer) error {
	tree := Tree{
		filepath.Join(landmarks.WorkspaceRoot(), ".timeless/candidates/", sagaName.String()),
	}
	return tree.SaveModuleLineage(modName, api.Lineage{
		Name: modName,
		Releases: []api.Release{
			{
				Name:  "candidate",
				Items: content,
			},
		},
	})
}

// Dependent builds will need to be *evicted* from the saga if a module
// is rebuilt and comes up with a different set of result contents.
// Not sure what a good UX is for that.  Only comes up in manual mode.
// Possible that we should save import resolutions to make this possible
// to recheck idempotently (we wouldn't want insanity to result from
// killing the reach process during that eviction phase!).

func SaveCandidateReplay(landmarks layout.Workspace, sagaName SagaName, modName api.ModuleName, mod api.Module, stderr io.Writer) error {
	tree := Tree{
		filepath.Join(landmarks.WorkspaceRoot(), ".timeless/candidates/", sagaName.String()),
	}

	// Rewrite ingests
	//  Error if export missing
	// TODO

	// Write rewritten module to file
	//  n.b. there's a good chance this still has "candidate" releases in it;
	//   that's okay for now, but we'll rewrite them before committing the saga.
	// TODO there's no func for this on Tree yet
	_ = tree

	return nil
}

// Seems like it would be nice to generate a viable 'mirrors.tl' file as well.
// But it's hard to spec out what that should behave like.
// Often, the locally viable URLs are *not* publicly viable.
