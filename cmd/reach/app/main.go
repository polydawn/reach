package reach

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"

	api "go.polydawn.net/go-timeless-api"
	catalogApp "go.polydawn.net/reach/app/catalog"
	ciApp "go.polydawn.net/reach/app/ci"
	emergeApp "go.polydawn.net/reach/app/emerge"
	waresApp "go.polydawn.net/reach/app/wares"
	"go.polydawn.net/reach/gadgets/catalog"
	"go.polydawn.net/reach/gadgets/layout"
	"go.polydawn.net/reach/gadgets/module"
	"go.polydawn.net/reach/gadgets/workspace"
)

func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) (exitCode int) {
	app := &cli.App{
		Name: "reach",
		UsageText: "Reach is a multipurpose tool for driving and managing Timeless Stack projects.\n" +
			"   Major functions of `reach` include:\n" +
			"\n" +
			"     - evaluating modules, which run pipelines of repeatr operations;\n" +
			"     - staging release candidates of the results;\n" +
			"     - commissioning entire generations of builds from many modules;\n" +
			"     - and publishing and managing release catalogs for distribution.\n" +
			"\n" +
			"   See https://repeatr.io/ for more complete documention!",
		Writer: stderr,

		// Must configure this to override an os.Exit(3).
		CommandNotFound: func(ctx *cli.Context, command string) {
			exitCode = 1
			fmt.Fprintf(stderr, "reach: incorrect usage: '%s' is not a %s subcommand\n", command, ctx.App.Name)
		},
	}

	app.Commands = append(app.Commands, cli.Command{
		Name:  "emerge",
		Usage: "evaluate a pipeline, logging intermediate results and reporting final exports",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "r, recursive",
				Usage: "if set, module evaluation will allow recursion: imports of candidate releases -- e.g., of the form \"catalog:$module:candidate:$item\" -- will cause that module to be freshly built rather than using an existing release.",
			},
		},
		Action: func(args *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			// Parse other flags.
			sn, _ := catalog.ParseSagaName("default") // TODO more complicated defaults and flags

			// Find workspace.
			workspaceLayout, err := layout.FindWorkspace(cwd)
			if err != nil {
				return err
			}
			workspace := workspace.Workspace{*workspaceLayout}

			// Behavior switch based on whether or not recursion is allowed.
			//  Future: unify these more...
			if args.Bool("recursive") {
				// Flip args into module names
				//  FIXME: anon invocation not sensibly handled here
				//  TODO: a lot of things not sensibly handled here...
				//   we should probably do something not unlike what the go commands do:
				//    - if it starts with a `./`, interpret as a path.
				//    - if it doesn't, interpret as a module name.
				//    - (and make some library functions that consistently do this for us.)
				moduleNames := []api.ModuleName(nil)
				for _, arg := range args.Args() {
					moduleNames = append(moduleNames, api.ModuleName(arg))
				}

				// Go!
				return emergeApp.EmergeMulti(workspace, moduleNames, *sn, stdout, stderr)
			} else {
				// Find (or expect) module (depending on args style).
				//  The arg is expected to be a *path* (not a module name
				//   (although the two are currently the same in practice)).
				var moduleLayout *layout.Module
				switch args.NArg() {
				case 0:
					moduleLayout, err = layout.FindModule(*workspaceLayout, cwd)
				case 1:
					// TODO: difference between moduleName and path also should be considered here (see other TODOs above).
					moduleLayout, err = layout.ExpectModule(*workspaceLayout, filepath.Join(cwd, args.Args()[0]))
				default:
					return fmt.Errorf("'reach emerge' takes zero or one args")
				}
				if err != nil {
					return err // oh no
				}

				// Load module.
				mod, err := module.Load(*moduleLayout)
				if err != nil {
					return fmt.Errorf("error loading module: %s", err)
				}

				// Go!
				return emergeApp.EvalModule(workspace, *moduleLayout, sn, *mod, stdout, stderr)
			}
		},
	})

	app.Commands = append(app.Commands, cli.Command{
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
				return fmt.Errorf("'reach ci' takes zero or one args")
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
	})

	app.Commands = append(app.Commands, cli.Command{
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
						return fmt.Errorf("'reach catalog lint' takes zero or one args")
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
	})

	app.Commands = append(app.Commands, cli.Command{
		Name:  "wares",
		Usage: "look up wares by release or candidate",
		Subcommands: []cli.Command{
			{
				Name:  "select",
				Usage: "View and search for wares",
				Subcommands: []cli.Command{
					{
						Name:      "candidates",
						Usage:     "List release candidates",
						ArgsUsage: "[<moduleNameOrPath> [<item-name>]]",
						Action: func(args *cli.Context) error {
							cwd, err := os.Getwd()
							if err != nil {
								return err
							}

							// Parse other flags.
							sn, _ := catalog.ParseSagaName("default") // TODO more complicated defaults and flags

							// Find workspace.
							workspaceLayout, err := layout.FindWorkspace(cwd)
							if err != nil {
								return err
							}

							ws := workspace.Workspace{*workspaceLayout}

							var itemName *api.ItemName
							var modNameOrPath string
							var modName *api.ModuleName
							switch args.NArg() {
							case 2:
								tmp := api.ItemName(args.Args()[1])
								itemName = &tmp
								fallthrough

							case 1:
								modNameOrPath = args.Args()[0]
								fallthrough
							case 0:
								modName, err = ModuleNameOrPath(ws, modNameOrPath, cwd)
								if err != nil {
									return err
								}
							default:
								return fmt.Errorf("select takes 0 or 1 item name")
							}

							return waresApp.ListCandidates(ws, *modName, *sn, itemName, stdout, stderr)
						},
					},
					{
						Name:      "releases",
						Usage:     "List releases",
						ArgsUsage: "[<moduleNameOrPath> [<releaseName> [<itemName>]]]",
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

							workspace := workspace.Workspace{*workspaceLayout}
							var modName *api.ModuleName
							var releaseName *api.ReleaseName
							var itemName *api.ItemName
							var modNameStr string
							switch args.NArg() {
							case 3:
								tmp := api.ItemName(args.Args()[2])
								itemName = &tmp
								fallthrough
							case 2:
								tmp := api.ReleaseName(args.Args()[1])
								releaseName = &tmp
								fallthrough
							case 1:
								modNameStr = args.Args()[0]
								fallthrough
							case 0:
								modName, err = ModuleNameOrPath(workspace, modNameStr, cwd)
								if err != nil {
									return err
								}
							default:
								return fmt.Errorf("select takes 0 or 1 item name")
							}
							return waresApp.ListReleases(workspace, *modName, releaseName, itemName, stdout, stderr)
						},
					},
				},
			},
			{
				Name:  "unpack",
				Usage: "Unpack wares",
				Subcommands: []cli.Command{
					{
						Name:      "wareID",
						Usage:     "Unpack a WareID to a path",
						ArgsUsage: "<wareId> [<outputPath>]",
						Action: func(args *cli.Context) error {
							unpackDir := "tmp.unpack"
							var wareId api.WareID
							switch args.NArg() {
							case 2:
								path := args.Args()[1]
								if path == "." {
									return fmt.Errorf("Won't unpack to '.', that would destroy your current directory")
								}
								unpackDir = path
								fallthrough
							case 1:
								var err error
								wareId, err = api.ParseWareID(args.Args()[0])
								if err != nil {
									return err
								}
							case 0:
								return fmt.Errorf("WareId is a required parameter.  See -h for help")
							default:
								return fmt.Errorf("unpack takes 1 or 2 arguments.")
							}
							cwd, err := os.Getwd()
							if err != nil {
								return err
							}

							// Find workspace.
							workspaceLayout, err := layout.FindWorkspace(cwd)
							if err != nil {
								return err
							}

							workspace := workspace.Workspace{*workspaceLayout}
							unpackPath := filepath.Join(cwd, unpackDir)
							err = os.MkdirAll(unpackPath, 0644)
							if err != nil {
								return err
							}
							return waresApp.UnpackWareID(ctx, workspace, wareId, unpackPath, stdout, stderr)
						},
					},
					{
						Name:      "candidate",
						Usage:     "Unpack a release candidate",
						ArgsUsage: "<moduleNameOrPath> <itemName> [<outputPath>]",
						Action: func(args *cli.Context) error {
							unpackDir := "tmp.unpack"
							cwd, err := os.Getwd()
							if err != nil {
								return err
							}

							// Find workspace.
							workspaceLayout, err := layout.FindWorkspace(cwd)
							if err != nil {
								return err
							}
							sn, _ := catalog.ParseSagaName("default") // TODO more complicated defaults and flags
							ws := workspace.Workspace{*workspaceLayout}
							switch args.NArg() {
							case 3:
								unpackDir = args.Args()[2]
								fallthrough
							case 2:
								modName, err := ModuleNameOrPath(ws, args.Args()[0], cwd)
								if err != nil {
									return err
								}
								itemName := api.ItemName(args.Args()[1])
								return waresApp.UnpackCandidate(ctx, ws, *sn, *modName, itemName, unpackDir, stdout, stderr)
							default:
								return fmt.Errorf("'unpack candidate' takes either 2 or 3 arguments.  See -h for details.")
							}
						},
					},
					{
						Name:      "release",
						Usage:     "Unpack a release ",
						ArgsUsage: "<moduleNameOrPath> <releaseName> <itemName> [<outputPath>]",
						Action: func(args *cli.Context) error {
							unpackDir := "tmp.unpack"
							cwd, err := os.Getwd()
							if err != nil {
								return err
							}

							// Find workspace.
							workspaceLayout, err := layout.FindWorkspace(cwd)
							if err != nil {
								return err
							}
							ws := workspace.Workspace{*workspaceLayout}
							switch args.NArg() {
							case 4:
								unpackDir = args.Args()[3]
								fallthrough
							case 3:
								modName, err := ModuleNameOrPath(ws, args.Args()[0], cwd)
								if err != nil {
									return err
								}
								releaseName := api.ReleaseName(args.Args()[1])
								itemName := api.ItemName(args.Args()[2])
								return waresApp.UnpackRelease(ctx, ws, *modName, releaseName, itemName, unpackDir, stdout, stderr)
							default:
								return fmt.Errorf("'unpack release' takes either 3 or 4 arguments.  See -h for details.")
							}
						},
					},
				},
			},
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name:  "synopsis",
		Usage: "list every command and subcommand, for quick reference",
		Action: func(args *cli.Context) error {
			printSynopsis(stderr, []string{"reach"}, app.Commands)
			return nil
		},
	})

	if err := app.Run(args); err != nil {
		exitCode = 1
		fmt.Fprintf(stderr, "reach: %s\n", err)
	}
	return
}

func printSynopsis(stderr io.Writer, stack []string, cmds []cli.Command) {
	for _, cmd := range cmds {
		if cmd.Subcommands != nil {
			printSynopsis(stderr, append(stack, cmd.Name), cmd.Subcommands)
		}
		if cmd.Action != nil {
			fmt.Fprintf(stderr, "%s\n", strings.Join(append(stack, cmd.Name), " "))
		}
	}
}
