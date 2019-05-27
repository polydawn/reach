package operation

import (
	"context"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/repeatr"
)

// PreparedOperation is a struct gathering all the results of resolving
// an api.Operation into an api.Formula: the formula itself, the context
// and URLs, and also the information for mapping any of the results
// back into locally scoped names from the api.Operation.
type PreparedOperation struct {
	Formula          api.Formula
	FormulaCtx       repeatr.FormulaContext
	OutputReverseMap map[api.AbsPath]api.SlotName
}

// Eval evaluates an PreparedOperation (derived from an api.Operation by
// the Resolve func, which does all name resolution) using a repeatr.RunFunc,
// and returns the results as an api.OperationRecord.
//
// The Eval and Resolve funcs are separated in this way so that code calling
// them can pause between the two functions and has the opportunity to log the
// prepared formulas, etc.
//
// Note: the Monitor system is not translated, and messages through that system
// will use paths as repeatr sees them, rather than slotrefs and slotnames.
//
func Eval(
	ctx context.Context,
	runTool repeatr.RunFunc, // the repeatr API to drive
	prop PreparedOperation, // The prepared operation -- use Resolve to generate this
	input repeatr.InputControl, // Optionally: input control.  The zero struct is no input (which is fine).
	monitor repeatr.Monitor, // Optionally: callbacks for progress monitoring.  Also where stdout/stderr is gathered.
) (*api.OperationRecord, error) {
	// The PreparedOperation should already be complete.  Run!
	runRecord, err := runTool(
		ctx,
		prop.Formula,
		prop.FormulaCtx,
		input,
		monitor,
	)
	if err != nil {
		return nil, err
	}

	// Convert outputs back to indexed by the slotnames we started with.
	opRecord := api.OperationRecord{
		FormulaRunRecord: *runRecord,
		Results:          make(map[api.SlotName]api.WareID, len(runRecord.Results)),
	}
	for pth, wareID := range runRecord.Results {
		opRecord.Results[prop.OutputReverseMap[pth]] = wareID
	}

	// Ya ta!
	return &opRecord, nil
}
