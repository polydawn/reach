package module

import (
	"context"
	"fmt"
	"os"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr"
	"go.polydawn.net/go-timeless-api/repeatr/fmt"
)

// FUTURE: would be nice to have each step eval return futures, and then
//  have a pool of 'worker' goroutines so we can have a '-j' param for actual run.
//  Of course, it's hard to say what's sensible other than '-j=9999' once you
//  throw e.g. a kubernetes cluster at it as a resource.  Threads are not the
//  real resource to watch -- just a simpleton correlate.

func Evaluate(
	ctx context.Context,
	mod api.Module,
	order []api.SubmoduleStepRef,
	pins map[api.SubmoduleSlotRef]api.WareID,
	wareSourcing api.WareSourcing,
	wareStaging api.WareStaging,
	runTool repeatr.RunFunc,
) (_ map[api.ItemName]api.WareID, err error) {
	return evaluate(ctx, "", mod, order, map[api.SlotRef]api.WareID{}, pins, wareSourcing, wareStaging, runTool)
}

func evaluate(
	ctx context.Context,
	ctxPth api.SubmoduleRef,
	mod api.Module,
	order funcs.StepTree,
	parentScope map[api.SlotRef]api.WareID,
	pins funcs.Pins,
	wareSourcing api.WareSourcing,
	wareStaging api.WareStaging,
	runTool repeatr.RunFunc,
) (_ map[api.ItemName]api.WareID, err error) {
	// Initialize map of locally scoped inputs.
	scope := map[api.SlotRef]api.WareID{}
	var ok bool
	for slotName, importRef := range mod.Imports {
		switch ref2 := importRef.(type) {
		case api.ImportRef_Catalog: // catalog references should already be resolved and handed to us in the pins map.
			scope[api.SlotRef{"", slotName}], ok = pins[api.SubmoduleSlotRef{"", api.SlotRef{"", slotName}}]
			if !ok {
				return nil, fmt.Errorf("missing pin for import %q in module %s", slotName, ctxPth)
			}
		case api.ImportRef_Parent: // parent references pluck something out of the parent scope.
			scope[api.SlotRef{"", slotName}], ok = parentScope[api.SlotRef(ref2)]
			if !ok {
				return nil, fmt.Errorf("missing pin for import %q in module %s", slotName, ctxPth)
			}
		case api.ImportRef_Ingest: // ingest references should *also* already be resolved and handed to us in the pins map.
			scope[api.SlotRef{"", slotName}], ok = pins[api.SubmoduleSlotRef{"", api.SlotRef{"", slotName}}]
			if !ok {
				return nil, fmt.Errorf("missing pin for import %q in module %s", slotName, ctxPth)
			}
		}
	}
	// Loop over steps at this level.  Append scope map with each result.
	for _, submStepRef := range order {
		if submStepRef.SubmoduleRef != "" {
			continue // belongs to a deeper level, handled by recursion already
		}
		fmt.Printf("steppin %v: %v\n", ctxPth, submStepRef)
		switch step := mod.Steps[submStepRef.StepName].(type) {
		case api.Operation:
			mon, monWaitCh := repeatrfmt.ServeMonitor(repeatrfmt.NewAnsiPrinter(os.Stdout, os.Stderr))
			record, err := repeatr.RunOperation(
				ctx,
				runTool,
				step,
				scope,
				wareSourcing,
				wareStaging,
				repeatr.InputControl{}, // input control is always zero for build jobs.
				mon,
			)
			close(mon.Chan)
			<-monWaitCh
			if err != nil {
				return nil, fmt.Errorf("failed evaluating operation %q: %s", submStepRef.Contextualize(ctxPth), err)
			}
			for slotName := range step.Outputs {
				scope[api.SlotRef{submStepRef.StepName, slotName}] = record.Results[slotName]
			}
			if record.ExitCode != 0 {
				// REVIEW: for the module, as with the operation alone: job unsuccessful isn't the same as eval error; make third return value?
				return nil, fmt.Errorf("operation %q exit code %d -- eval halted", submStepRef.Contextualize(ctxPth), record.ExitCode)
			}
		case api.Module:
			submoduleResults, err := evaluate(
				ctx,
				ctxPth.Child(submStepRef.StepName),
				step,
				order.DetachSubtree(submStepRef.StepName),
				scope,
				pins.DetachSubtree(submStepRef.StepName),
				wareSourcing,
				wareStaging,
				runTool,
			)
			if err != nil {
				return nil, err
			}
			for itemName := range step.Exports {
				scope[api.SlotRef{submStepRef.StepName, api.SlotName(itemName)}] = submoduleResults[itemName]
			}
		case nil:
			panic("order incongruity")
		default:
			panic("unreachable")
		}
	}
	// Extract exports from local scope and return them under their export names.
	exportedResults := make(map[api.ItemName]api.WareID, len(mod.Exports))
	for exportName, slotRef := range mod.Exports {
		pin, ok := scope[slotRef]
		if !ok {
			panic(fmt.Errorf("module %q tries to use %s as an export but it is not in scope", ctxPth, slotRef))
		}
		exportedResults[exportName] = pin
	}
	return exportedResults, nil
}
