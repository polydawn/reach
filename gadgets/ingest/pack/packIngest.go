package packingest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/go-timeless-api/rio/client/exec"
)

type Config struct {
	StagingArea api.WareStaging
}

func (cfg Config) Resolve(ctx context.Context, ingestRef api.ImportRef_Ingest) (
	*api.WareID,
	*api.WareSourcing,
	error,
) {
	// Args handling.
	if ingestRef.IngestKind != "pack" {
		return nil, nil, fmt.Errorf("pack ingest: invalid args: ingest ref must start with \"ingest:pack:\"")
	}
	refArgsHunks := strings.SplitN(ingestRef.Args, ":", 2)
	if len(refArgsHunks) != 2 {
		return nil, nil, fmt.Errorf("pack ingest: invalid args: need a pack type (e.g. \"tar\") and a path, separated by a colon (ex: \"ingest:pack:tar:./here\")")
	}
	packType := api.PackType(refArgsHunks[0])
	pth := refArgsHunks[1]
	// future: it's possible we'll want to parse further, maybe pack opts in parens after packtype

	// Absolutize path asap.
	//  We're perfectly happy to work with relative paths as ingest params,
	//  but it's a mess of unpleasantness to log and debug if we carry them.
	pth, err := filepath.Abs(pth)
	if err != nil {
		return nil, nil, fmt.Errorf("catastrophe, cannot find cwd: %s", err)
	}

	// Pick a single place where we're going to store output.
	warehouse := cfg.StagingArea.ByPackType[packType]

	// Apply rio.
	wareID, err := rioclient.PackFunc(
		ctx,
		packType,
		pth,
		api.FilesetPackFilter_Flatten,
		warehouse,
		rio.Monitor{},
	)

	// Return the wareID we got from packing, and waresourcing can just be
	// precisely the one we picked out to use already.
	wareSourcing := &api.WareSourcing{}
	wareSourcing.AppendByWare(wareID, warehouse)
	return &wareID, wareSourcing, err
}
