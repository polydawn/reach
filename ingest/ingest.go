package ingest

import (
	"context"
	"fmt"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/stellar/ingest/git"
)

func Resolve(ctx context.Context, ingestRef api.ImportRef_Ingest) (*api.WareID, *api.WareSourcing, error) {
	switch ingestRef.IngestKind {
	case "git":
		return gitingest.Resolve(ctx, ingestRef)
	default:
		return nil, nil, fmt.Errorf("ingest: kind %q not known", ingestRef.IngestKind)
	}
}
