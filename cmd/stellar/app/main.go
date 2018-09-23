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
					pth := ""
					switch args.NArg() {
					case 0:
						cwd, err := os.Getwd()
						if err != nil {
							return err
						}
						pth = cwd
					case 1:
						pth = args.Args()[0]
					default:
						return fmt.Errorf("'stellar emerge' takes zero or one args")
					}

					landmarks, err := layout.FindLandmarks(pth)
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
					return emergeApp.EvalModule(*landmarks, *mod, stdout, stderr)
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
					return ciApp.Loop(*landmarks, *mod, stdout, stderr)
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
								landmarks, err := layout.FindLandmarks(cwd)
								if err != nil {
									return err
								}
								if landmarks.ModuleCatalogRoot == "" {
									return fmt.Errorf("no catalog found")
								}
								pth = landmarks.ModuleCatalogRoot
							case 1:
								var err error
								pth, err = filepath.Abs(args.Args()[0])
								if err != nil {
									panic(err)
								}
								fi, err := os.Stat(pth)
								if err != nil {
									return fmt.Errorf("'stellar catalog lint' should be aimed at a directory: %s", err)
								}
								if fi.Mode()&os.ModeType != os.ModeDir {
									return fmt.Errorf("'stellar catalog lint' should be aimed at a directory")
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
