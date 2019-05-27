package operation

import (
	"fmt"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/repeatr"
)

// Resolve takes an api.Operation, the local scoped names, and all applicable
// ware sourcing and staging config, and turns them into a PreparedOperation.
//
// The returned PreparedOperation object has enough information to be
// given to repeatr.RunFunc, plus additional information for how to map
// the repeatr RunRecord back into the local-scope names of the Operation's
// outputs.  The Eval function in this package will take a PreparedOperation
// and evaluate it, and handle the mapping back into api.OperationRecord.
//
// ---
//
// There's a fair amount of shuffling inside this function.  Some context:
//
// Formulas index everything by paths, because paths are material to the
// outcome of the computation, and slot names as used in Operations *are not*.
// This makes Formulas, when hashed, a useful "primary key" for other lookups;
// but it means we need to pivot some things to go from Operation to Formula.
//
// The warehousing parameters of this function are similarly more opinionated
// than the level of detail that the RunFunc interface allows; since
// operation.Resolve is intended to be used when doing lots of operations, it
// makes more sense to use WareSourcing and WareStaging configuration.
// For the same reason, it makes sense to steer toward output warehouses
// that are the most reusable: thus, although the repeatr layer supports a
// wider set of options, here, only content-addressable warehouses, indexed by
// packType, are allowed as WareStaging arguments.
//
func Resolve(
	op api.Operation, // What protype of a formula to bind and run.
	scope map[api.SlotRef]api.WareID, // What slots are in scope to reference as inputs.
	wareSourcing api.WareSourcing, // Suggestions on where to get wares.
	wareStaging api.WareStaging, // Instructions on where to store output wares.
) (*PreparedOperation, error) {
	// Initialize everything we're about to fill in.
	//  (The Formula.Action just comes along, unchanged.)
	prop := &PreparedOperation{
		api.Formula{
			Inputs:  make(map[api.AbsPath]api.WareID, len(op.Inputs)),
			Action:  op.Action,
			Outputs: make(map[api.AbsPath]api.FormulaOutputSpec, len(op.Outputs)),
		},
		repeatr.FormulaContext{
			FetchUrls: make(map[api.AbsPath][]api.WarehouseLocation),
			SaveUrls:  make(map[api.AbsPath]api.WarehouseLocation),
		},
		make(map[api.AbsPath]api.SlotName, len(op.Outputs)),
	}

	// Resolve inputs to WareID.
	for pth, slotRef := range op.Inputs {
		pin, ok := scope[slotRef]
		if !ok {
			return nil, fmt.Errorf("cannot provide an input ware for slotref %q: no such ref in scope", slotRef)
		}
		prop.Formula.Inputs[pth] = pin
	}

	// Fill in outputs in Repeatr format.
	//  Save the reverse mappings back into slotnames as we go; we'll need these later.
	for slotName, pth := range op.Outputs {
		prop.Formula.Outputs[pth] = api.FormulaOutputSpec{
			PackType: "tar", // FUTURE: the api.Operation layer doesn't really support config for this yet, but should
			// FUTURE: filters are also conspicuously missing at the api.Operation layer, and should be included
		}
		prop.OutputReverseMap[pth] = slotName
	}

	// Populate FormulaContext with only the relevant FetchURLs from WareSourcing.
	wareSourcing = wareSourcing.PivotToInputs(prop.Formula)
	for pth, wareID := range prop.Formula.Inputs {
		prop.FormulaCtx.FetchUrls[pth] = wareSourcing.ByWare[wareID]
	}

	// Populate FormulaContext.SaveURLs according to WareStaging.
	for pth, outSpec := range prop.Formula.Outputs {
		prop.FormulaCtx.SaveUrls[pth] = wareStaging.ByPackType[outSpec.PackType]
	}

	// All flipped.  Return!
	return prop, nil
}
