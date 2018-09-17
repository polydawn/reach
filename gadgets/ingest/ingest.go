package ingest

import (
	"context"
	"fmt"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/stellar/gadgets/ingest/git"
	"go.polydawn.net/stellar/gadgets/ingest/literal"
	"go.polydawn.net/stellar/gadgets/ingest/pack"
)

type Config struct {
	StagingArea api.WareStaging
}

func (cfg Config) Resolve(ctx context.Context, ingestRef api.ImportRef_Ingest) (*api.WareID, *api.WareSourcing, error) {
	switch ingestRef.IngestKind {
	case "git":
		return gitingest.Resolve(ctx, ingestRef)
	case "pack":
		return packingest.Config{cfg.StagingArea}.Resolve(ctx, ingestRef)
	case "literal":
		return literalingest.Resolve(ctx, ingestRef)
	default:
		return nil, nil, fmt.Errorf("ingest: kind %q not known", ingestRef.IngestKind)
	}
}
