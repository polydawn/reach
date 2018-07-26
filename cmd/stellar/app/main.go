package stellar

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/stellar/catalog"
	"go.polydawn.net/stellar/ingest/git"
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
					return evalModule(*landmarks, *mod, stdout, stderr)
				},
			},
			{
				Name:  "ci",
				Usage: "given a module with one ingest using git, build it once, then build it again each time the git repo updates",
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
					var hingeIngest api.ImportRef_Ingest
					for _, imp := range mod.Imports {
						switch imp2 := imp.(type) {
						case api.ImportRef_Ingest:
							switch imp2.IngestKind {
							case "git":
								if hingeIngest != (api.ImportRef_Ingest{}) {
									return fmt.Errorf("a module for use in CI mode can only have one ingest!")
								}
								hingeIngest = imp2
							default:
								return fmt.Errorf("a module for use in CI mode can only have one ingest, and it must be 'ingest:git'")
							}
						}
					}
					if hingeIngest == (api.ImportRef_Ingest{}) {
						return fmt.Errorf("a module for use in CI mode must have one ingest, and it must be 'ingest:git'")
					}
					previouslyIngested := api.WareID{}
					for {
						newlyIngested, _, err := gitingest.Resolve(context.Background(), hingeIngest)
						if err != nil {
							return err
						}
						if *newlyIngested == previouslyIngested {
							time.Sleep(1260 * time.Millisecond)
							continue
						}
						fmt.Fprintf(stderr, "found new git hash!  evaluating %s\n", newlyIngested)
						if err := evalModule(*landmarks, *mod, stdout, stderr); err != nil {
							return err
						}
						fmt.Fprintf(stderr, "CI execution done, successfully.  Going into standby until more changes.\n")
						previouslyIngested = *newlyIngested
					}
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
