package stellar

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/urfave/cli"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/catalog"
	"go.polydawn.net/stellar/catalog/hitch"
	"go.polydawn.net/stellar/ingest"
	"go.polydawn.net/stellar/layout"
	"go.polydawn.net/stellar/module"
)

func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) (exitCode int) {
	app := &cli.App{
		Name:      "stellar",
		Usage:     "sidereal repeatr",
		UsageText: "Stellar builds modules of repeatr operations, stages releases of the results, and can commission builds of entire generations of atomic releases from many modules.",
		Writer:    stderr,
		Commands: []cli.Command{
			{
				Name:  "emerge",
				Usage: "evaluate a pipeline, logging intermediate results and reporting final exports",
				Action: func(ctx *cli.Context) error {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					landmarks, err := layout.FindLandmarks(cwd)
					if err != nil {
						return err
					}
					if landmarks.ModuleRoot == "" {
						return fmt.Errorf("no module found -- run this command in a module dir (e.g contains module.tl file), or specify a path to one")
					}
					mod, err := module.Load(*landmarks)
					if err != nil {
						return fmt.Errorf("error loading module: %s", err)
					}
					{
						fmt.Fprintf(stderr, "module loaded\n")
						ord, err := funcs.ModuleOrderStepsDeep(*mod)
						if err != nil {
							return err
						}
						fmt.Fprintf(stderr, "module contains %d steps\n", len(ord))
						fmt.Fprintf(stderr, "module evaluation plan order:\n")
						for i, fullStepRef := range ord {
							fmt.Fprintf(stderr, "  - %.2d: %s\n", i+1, fullStepRef)
						}
						wareStaging := api.WareStaging{ // FIXME needs to be defined by workspace
							ByPackType: map[api.PackType]api.WarehouseLocation{"tar": "ca+file://.timeless/warehouse/"},
						}
						wareSourcing := api.WareSourcing{}
						wareSourcing.AppendByPackType("tar", "ca+file://.timeless/warehouse/") // FIXME needs to be defined by workspace
						catalogHandle := hitch.FSCatalog{[]catalog.Tree{
							{landmarks.ModuleCatalogRoot},
							{filepath.Join(landmarks.WorkspaceRoot, ".timeless/catalogs/upstream")}, // TODO fix hardcoded "upstream" param
						}}
						pins, pinWs, err := funcs.ResolvePins(*mod, catalogHandle.ViewCatalog, catalogHandle.ViewWarehouses, ingest.Resolve)
						if err != nil {
							return err
						}
						wareSourcing.Append(*pinWs)
						fmt.Fprintf(stderr, "imports pinned to hashes:\n")
						allSlotRefs := []api.SubmoduleSlotRef{}
						for k, _ := range pins {
							allSlotRefs = append(allSlotRefs, k)
						}
						sort.Sort(api.SubmoduleSlotRefList(allSlotRefs))
						for _, k := range allSlotRefs {
							fmt.Fprintf(stderr, "  - %q: %s\n", k, pins[k])
						}
						// step step step!
						exports, err := module.Evaluate(
							context.Background(),
							*mod,
							ord,
							pins,
							wareSourcing,
							wareStaging,
							repeatrclient.Run,
						)
						if err != nil {
							return fmt.Errorf("evaluating module: %s", err)
						}
						fmt.Fprintf(stderr, "module eval complete.\n")
						fmt.Fprintf(stderr, "module exports:\n")
						for k, v := range exports {
							fmt.Fprintf(stderr, "  - %q: %v\n", k, v)
						}
					}
					return nil
				},
			},
			{
				Name:  "catalog",
				Usage: "catalog subcommands help maintain the release catalog info tree",
				Subcommands: []cli.Command{
					{
						Name:  "lint",
						Usage: "verify the entire catalog tree is in canonical form (rewrites all files)",
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "rewrite",
								Usage: "if set, all files will be rewritten to ensure bytewise canonicalization",
							},
						},
						Action: func(ctx *cli.Context) error {
							cwd, err := os.Getwd()
							if err != nil {
								return err
							}
							landmarks, err := layout.FindLandmarks(cwd)
							if err != nil {
								return err
							}
							if landmarks.ModuleCatalogRoot == "" {
								return fmt.Errorf("no catalog found")
							}
							warnings := 0
							err = catalog.Linter{
								Tree: catalog.Tree{landmarks.ModuleCatalogRoot}, // TODO as the name implies, this isn't generalized enough.  shouldn't be just modulecatalogs that are supported.
								WarnBehavior: func(msg string, _ func()) {
									warnings++
									fmt.Fprintf(stderr, "WARN: %s\n", msg)
								},
								Rewrite: ctx.Bool("rewrite"),
							}.Lint()
							fmt.Fprintf(stderr, "%d total warnings\n", warnings)
							if warnings > 0 {
								exitCode = 4 // TODO standardize
							}
							return err
						},
					},
				},
			},
		},
		// Must configure this to override an os.Exit(3).
		CommandNotFound: func(ctx *cli.Context, command string) {
			exitCode = 1
			fmt.Fprintf(stderr, "stellar: incorrect usage: '%s' is not a %s subcommand\n", command, ctx.App.Name)
		},
	}
	if err := app.Run(args); err != nil {
		exitCode = 1
		fmt.Fprintf(stderr, "stellar: %s\n", err)
	}
	return
}
