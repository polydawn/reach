package emergeApp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/gadgets/catalog"
	"go.polydawn.net/stellar/gadgets/catalog/hitch"
	"go.polydawn.net/stellar/gadgets/ingest"
	"go.polydawn.net/stellar/gadgets/layout"
	"go.polydawn.net/stellar/gadgets/module"
)

func EvalModule(landmarks layout.Landmarks, mod api.Module, stdout, stderr io.Writer) error {
	fmt.Fprintf(stderr, "module loaded\n")
	ord, err := funcs.ModuleOrderStepsDeep(mod)
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "module contains %d steps\n", len(ord))
	fmt.Fprintf(stderr, "module evaluation plan order:\n")
	for i, fullStepRef := range ord {
		fmt.Fprintf(stderr, "  - %.2d: %s\n", i+1, fullStepRef)
	}
	wareStaging := api.WareStaging{ByPackType: map[api.PackType]api.WarehouseLocation{"tar": landmarks.StagingWarehouse}}
	wareSourcing := api.WareSourcing{}
	wareSourcing.AppendByPackType("tar", landmarks.StagingWarehouse)
	catalogHandle := hitch.FSCatalog{[]catalog.Tree{
		{landmarks.ModuleCatalogRoot},
		{filepath.Join(landmarks.WorkspaceRoot, ".timeless/catalogs/upstream")}, // TODO fix hardcoded "upstream" param
	}}
	resolveTool := ingest.Config{
		landmarks.ModuleRoot,
		wareStaging, // FUTURE: should probably use different warehouse for this, so it's easier to GC the shortlived objects
	}.Resolve
	pins, pinWs, err := funcs.ResolvePins(mod, catalogHandle.ViewCatalog, catalogHandle.ViewWarehouses, resolveTool)
	if err != nil {
		return err
	}
	wareSourcing.Append(*pinWs)
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
	os.Setenv("REPEATR_MEMODIR", filepath.Join(landmarks.WorkspaceRoot, ".timeless/memo"))
	exports, err := module.Evaluate(
		context.Background(),
		mod,
		ord,
		pins,
		wareSourcing,
		wareStaging,
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
	return nil
}
