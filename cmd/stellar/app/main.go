package stellar

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"github.com/urfave/cli"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/stellar/layout"
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
					f, err := os.Open(filepath.Join(ti.Root, "module.tl"))
					if err != nil {
						return err
					}
					mod := api.Module{}
					unm := refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, f, api.Atlas_Module)
					err = unm.Unmarshal(&mod)
					if err != nil {
						return err
					}
					fmt.Fprintf(stderr, "workspace loaded\n")
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
		fmt.Fprintf(stderr, "stellar: incorrect usage: %s", err)
	}
	return
}
