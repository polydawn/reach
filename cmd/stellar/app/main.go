package stellar

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/fmt"
	"go.polydawn.net/stellar/hitch"
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
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "print",
				Usage: "Define the format of logs to print.",
				// MoreUsage: "Define the format of logs to print.  The default is ansi and quite verbose."+
				//   "Other formats, e.g. \"json\" can be specified, "+
				//   "and certain parts of printing enabled or disabled by e.g. \"ansi-repeatrLogs\" "+
				//   "to remove operation setup and teardown logs, or \"ansi+repeatrOutput\" to print "+
				//   "the complete output of contained jobs.",
			},
		},
		Commands: []cli.Command{
			{
				Name:  "emerge",
				Usage: "todo docs",
				Action: func(ctx *cli.Context) error {
					p := parsePrintFlag(ctx.String("print"), stdout, stderr)

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
						p.PrintLog("workspace loaded\n")
						ord, err := funcs.ModuleOrderStepsDeep(*mod)
						if err != nil {
							return err
						}
						p.PrintLog(fmt.Sprintf("module contains %d steps\n", len(ord)))
						p.PrintLog(stellarfmt.Tmpl("module evaluation plan order:\n"+
							`{{range $k, $v := . -}}`+
							`{{printf "  - %.2d: %s\n" (inc $k) $v}}`+
							`{{end}}`, ord))
						wareSourcing := api.WareSourcing{}
						wareSourcing.AppendByPackType("tar", "ca+file://.timeless/warehouse/")
						catalogHandle := hitch.FSCatalog{ti.CatalogRoot}
						pins, pinWs, err := funcs.ResolvePins(*mod, catalogHandle.ViewCatalog, catalogHandle.ViewWarehouses, ingest.Resolve)
						if err != nil {
							return err
						}
						wareSourcing.Append(*pinWs)
						p.PrintLog(stellarfmt.Tmpl("imports pinned to hashes:\n"+
							`{{range $k, $v := . -}}`+
							`{{printf "  - %q: %s\n" $k $v}}`+
							`{{end}}`, pins))
						// step step step!
						exports, err := module.Evaluate(
							context.Background(),
							*mod,
							ord,
							pins,
							wareSourcing,
							repeatrclient.Run,
						)
						if err != nil {
							return fmt.Errorf("evaluating module: %s", err)
						}
						p.PrintLog("module eval complete.\n")
						p.PrintLog(stellarfmt.Tmpl("module exports:\n"+
							`{{range $k, $v := . -}}`+
							`{{printf "  - %q: %s\n" $k $v}}`+
							`{{end}}`, exports))
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

func parsePrintFlag(flags string, stdout, stderr io.Writer) (v stellarfmt.Printer) {
	v.PrintLog = stellarfmt.PrinterLogText{stderr}.PrintLog
	return
}
