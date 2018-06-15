package stellar

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli"
)

func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
					return nil
				},
			},
		},
	}
	if err := app.Run(args); err != nil {
		fmt.Fprintf(stderr, "Incorrect usage: %s", err)
		return 1
	}
	return 0
}
