package stellar

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/urfave/cli"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/hitch"
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
				Usage: "todo docs",
				Action: func(ctx *cli.Context) error {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					ti, err := layout.FindTree(cwd)
					if err != nil {
						return err
					}
					switch ti.Singleton {
					case true:
						mod, err := module.LoadByPath(*ti, "module.tl")
						if err != nil {
							return err
						}
						fmt.Fprintf(stderr, "workspace loaded\n")
						ord, err := funcs.ModuleOrderStepsDeep(*mod)
						if err != nil {
							return err
						}
						fmt.Fprintf(stderr, "module contains %d steps\n", len(ord))
						fmt.Fprintf(stderr, "module evaluation plan order:\n")
						for i, fullStepRef := range ord {
							fmt.Fprintf(stderr, "  - %.2d: %s\n", i+1, fullStepRef)
						}
						pins, err := funcs.ResolvePins(*mod, hitch.FSCatalog{ti.CatalogRoot}.ViewCatalog)
						if err != nil {
							return err
						}
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
					case false:
						panic("TODO")
					}
					return nil
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
		fmt.Fprintf(stderr, "stellar: %s", err)
	}
	return
}
