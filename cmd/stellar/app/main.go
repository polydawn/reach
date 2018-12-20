package stellar

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"go.polydawn.net/stellar/app/catalog"
	"go.polydawn.net/stellar/app/ci"
	"go.polydawn.net/stellar/app/emerge"
	"go.polydawn.net/stellar/gadgets/catalog"
	"go.polydawn.net/stellar/gadgets/layout"
	"go.polydawn.net/stellar/gadgets/module"
	"go.polydawn.net/stellar/gadgets/workspace"
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
				Action: func(args *cli.Context) error {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}

					// Find workspace.
					workspaceLayout, err := layout.FindWorkspace(cwd)
					if err != nil {
						return err
					}

					// Find (or expect) module (depending on args style).
					var moduleLayout *layout.Module
					switch args.NArg() {
					case 0:
						moduleLayout, err = layout.FindModule(*workspaceLayout, cwd)
					case 1:
						moduleLayout, err = layout.ExpectModule(*workspaceLayout, filepath.Join(cwd, args.Args()[0]))
					default:
						return fmt.Errorf("'stellar emerge' takes zero or one args")
					}
					if err != nil {
						return err
					}

					// Load module and load workspace conf.
					mod, err := module.Load(*moduleLayout)
					if err != nil {
						return fmt.Errorf("error loading module: %s", err)
					}
					workspace := workspace.Workspace{*workspaceLayout}

					// Parse other flags.
					sn, _ := catalog.ParseSagaName("hax") // TODO more complicated defaults and flags

					// Go!
					return emergeApp.EvalModule(workspace, *moduleLayout, sn, *mod, stdout, stderr)
				},
			},
			{
				Name:  "ci",
				Usage: "given a module with one ingest using git, build it once, then build it again each time the git repo updates",
				Action: func(args *cli.Context) error {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}

					// Find workspace.
					workspaceLayout, err := layout.FindWorkspace(cwd)
					if err != nil {
						return err
					}

					// Find (or expect) module (depending on args style).
					var moduleLayout *layout.Module
					switch args.NArg() {
					case 0:
						moduleLayout, err = layout.FindModule(*workspaceLayout, cwd)
					case 1:
						moduleLayout, err = layout.ExpectModule(*workspaceLayout, filepath.Join(cwd, args.Args()[0]))
					default:
						return fmt.Errorf("'stellar ci' takes zero or one args")
					}
					if err != nil {
						return err
					}

					// Load module and load workspace conf.
					mod, err := module.Load(*moduleLayout)
					if err != nil {
						return fmt.Errorf("error loading module: %s", err)
					}
					workspace := workspace.Workspace{*workspaceLayout}

					return ciApp.Loop(workspace, *moduleLayout, *mod, stdout, stderr)
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
						Action: func(args *cli.Context) error {
							pth := ""
							switch args.NArg() {
							case 0:
								cwd, err := os.Getwd()
								if err != nil {
									return err
								}
								workspaceLayout, err := layout.FindWorkspace(cwd)
								if err != nil {
									return err
								}
								pth = workspaceLayout.CatalogRoot()
							case 1:
								var err error
								pth, err = filepath.Abs(args.Args()[0])
								if err != nil {
									panic(err)
								}
							default:
								return fmt.Errorf("'stellar catalog lint' takes zero or one args")
							}

							warnings := 0
							err := catalogApp.Linter{
								Tree: catalog.Tree{pth},
								WarnBehavior: func(msg string, _ func()) {
									warnings++
									fmt.Fprintf(stderr, "WARN: %s\n", msg)
								},
								Rewrite: args.Bool("rewrite"),
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
