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
	ModuleDir   string
	StagingArea api.WareStaging
}

func (cfg Config) Resolve(ctx context.Context, ingestRef api.ImportRef_Ingest) (*api.WareID, *api.WareSourcing, error) {
	switch ingestRef.IngestKind {
	case "git":
		return gitingest.Config{
			ModuleDir: cfg.ModuleDir,
		}.Resolve(ctx, ingestRef)
	case "pack":
		return packingest.Config{
			ModuleDir:   cfg.ModuleDir,
			StagingArea: cfg.StagingArea,
		}.Resolve(ctx, ingestRef)
	case "literal":
		return literalingest.Resolve(ctx, ingestRef)
	default:
		return nil, nil, fmt.Errorf("ingest: kind %q not known", ingestRef.IngestKind)
	}
}
