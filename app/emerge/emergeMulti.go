package emergeApp

import (
	"fmt"
	"io"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/reach/gadgets/catalog"
	"go.polydawn.net/reach/gadgets/commission"
	"go.polydawn.net/reach/gadgets/module"
	"go.polydawn.net/reach/gadgets/workspace"
)

func EmergeMulti(
	ws workspace.Workspace, // needed for... everything.
	moduleNames []api.ModuleName, // list of modules by name that we def want eval'd.
	sagaName catalog.SagaName, // required so we can pass catalogs between modules.
	stdout, stderr io.Writer,
) error {
	fmt.Printf("asks: %v\n", moduleNames)
	order, err := commission.CommissionOrder(
		ws,
		moduleNames...,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("order found: %v\n", order)
	for _, modName := range order {
		modLayout := ws.GetModuleLayout(modName)
		mod, err := module.Load(*modLayout)
		if err != nil {
			panic(err)
		}
		err = EvalModule(
			ws,
			*modLayout,
			&sagaName,
			*mod,
			stdout, stderr,
		)
		if err != nil {
			panic(err)
		}
	}
	return nil
}
