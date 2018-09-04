package literalingest

import (
	"context"
	"fmt"

	"go.polydawn.net/go-timeless-api"
)

func Resolve(ctx context.Context, ingestRef api.ImportRef_Ingest) (
	*api.WareID,
	*api.WareSourcing,
	error,
) {
	// Args handling.
	if ingestRef.IngestKind != "literal" {
		return nil, nil, fmt.Errorf("literal ingest: invalid args: ingest ref must start with \"ingest:literal:\"")
	}
	wareID, err := api.ParseWareID(ingestRef.Args)
	if err != nil {
		return nil, nil, fmt.Errorf("literal ingest: invalid args: %s", err)
	}
	return &wareID, &api.WareSourcing{
	// Caveat Emptor!
	//
	// Literal ingests currently lack a way to express any recommended warehouse
	// addresses.
	//
	// This means they are going to be difficult to use unless you can supply
	// warehouse addresses through other mechanisms -- at present, having it
	// in any of the hardcoded default content-addressible warehouses is
	// sufficient; in the future there are more planned options e.g. suggesting
	// warehouses addresses to by used ByPackType via an env var.
	}, nil
}
