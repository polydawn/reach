package module

import (
	"context"
	"fmt"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/repeatr"
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
	runTool repeatr.RunFunc,
) (_ map[api.ItemName]api.WareID, err error) {
	return evaluate(ctx, "", mod, order, pins, runTool)
}

func evaluate(
	ctx context.Context,
	ctxPth api.SubmoduleRef,
	mod api.Module,
	order []api.SubmoduleStepRef,
	pins map[api.SubmoduleSlotRef]api.WareID,
	runTool repeatr.RunFunc,
) (_ map[api.ItemName]api.WareID, err error) {
	// Initialize map of locally scoped inputs.
	scope := map[api.SlotRef]api.WareID{}
	for submSlotRef, pin := range pins {
		if submSlotRef.SubmoduleRef != "" {
			continue // imported by a deeper level, not in our scope.
		}
		scope[submSlotRef.SlotRef] = pin
	}
	// Loop over steps at this level.  Append scope map with each result.
	for _, submStepRef := range order {
		if submStepRef.SubmoduleRef != "" {
			continue // belongs to a deeper level, handled by recursion already
		}
		switch step := mod.Steps[submStepRef.StepName].(type) {
		case api.Operation:
			boundOp := api.BoundOperation{
				InputPins: make(map[api.SlotRef]api.WareID),
				Operation: step,
			}
			for slotRef := range step.Inputs {
				pin, ok := scope[slotRef]
				if !ok {
					// This could be either because your order was not toposorted, or
					//  because of out-of-scope references.  Message could improve.
					panic(fmt.Errorf("operation %q tries to use %q but it is not in scope", submStepRef.Contextualize(ctxPth), slotRef))
				}
				boundOp.InputPins[slotRef] = pin
			}
			record, err := runTool(
				ctx,
				boundOp,
				api.WareSourcing{},     // FUTURE: move beyond placeholder... possible this should come along with pins.
				repeatr.InputControl{}, // input control is always zero for build jobs.
				repeatr.Monitor{},
			)
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
			panic("TODO submodule eval")
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
