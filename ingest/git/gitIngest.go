package gitingest

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"go.polydawn.net/go-timeless-api"
)

func Resolve(ctx context.Context, ingestRef api.ImportRef_Ingest) (
	*api.WareID,
	*api.WareSourcing,
	error,
) {
	// Args handling.
	if ingestRef.IngestKind != "git" {
		return nil, nil, fmt.Errorf("git ingest: invalid args: ingest ref must start with \"ingest:git:\"")
	}
	refArgsHunks := strings.SplitN(ingestRef.Args, ":", 2)
	if len(refArgsHunks) != 2 {
		return nil, nil, fmt.Errorf("git ingest: invalid args: need a path and a git ref (e.g. branch name), separated by a colon (ex: \"ingest:git:.:HEAD\")")
	}
	pth := refArgsHunks[0]
	gitRefName := plumbing.ReferenceName(refArgsHunks[1])

	// Open the repo.  (Currently we're only supporting local ones.)
	r, err := git.PlainOpen(pth)
	if err != nil {
		return nil, nil, err
	}
	// Flip references into a map.
	//  (If we have a symbolic ref, we'll have to go through at least twice anyway, so, might as well.)
	refsItr, err := r.References()
	if err != nil {
		return nil, nil, err
	}
	refs := map[plumbing.ReferenceName]*plumbing.Reference{}
	refsItr.ForEach(func(ref *plumbing.Reference) error {
		refs[ref.Name()] = ref
		return nil
	})

	// Do lookup, resolving symbolics as necessary.
	ref := refs[gitRefName]
	for i := 0; i < 10; i++ {
		switch ref.Type() {
		case plumbing.SymbolicReference:
			ref = refs[ref.Target()]
		case plumbing.HashReference:
			wareID := api.WareID{"git", ref.Hash().String()}
			ws := api.WareSourcing{}
			ws.AppendByWare(wareID, api.WarehouseLocation("file://"+pth))
			return &wareID, &ws, nil
		default:
			panic("git ingest: unknown ref type")
		}
	}
	return nil, nil, fmt.Errorf("git ingest: too many layers of symbolic reference")
}
