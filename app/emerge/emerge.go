package emergeApp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"github.com/polydawn/refmt/obj/atlas"
	"github.com/warpfork/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/gadgets/catalog"
	hitchGadget "go.polydawn.net/stellar/gadgets/catalog/hitch"
	"go.polydawn.net/stellar/gadgets/ingest"
	"go.polydawn.net/stellar/gadgets/layout"
	"go.polydawn.net/stellar/gadgets/module"
	"go.polydawn.net/stellar/gadgets/workspace"
)

func EvalModule(
	workspace workspace.Workspace, // needed to figure out if we have a moduleName.
	landmarks layout.Module, // needed in case of ingests with relative paths.
	sagaName *catalog.SagaName, // may have been provided as a flag.
	mod api.Module, // already helpfully loaded for us.
	stdout, stderr io.Writer,
) error {
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
	wareStaging := api.WareStaging{ByPackType: map[api.PackType]api.WarehouseLocation{"tar": landmarks.StagingWarehouseLoc()}}
	wareSourcing := api.WareSourcing{}
	wareSourcing.AppendByPackType("tar", landmarks.StagingWarehouseLoc())
	viewCatalogTool, viewWarehousesTool := hitchGadget.ViewTools([]catalog.Tree{
		// refactor note: we used to stack several catalog dirs here, but have backtracked from allowing that.
		// so it's possible there's a layer of abstraction here that should be removed outright; have not fully reviewed.
		{landmarks.CatalogRoot()},
	}...)
	if sagaName != nil {
		viewCatalogTool = hitchGadget.WithCandidates(
			viewCatalogTool,
			catalog.Tree{filepath.Join(landmarks.WorkspaceRoot(), ".timeless/candidates/", sagaName.String())},
		)
	}
	resolveTool := ingest.Config{
		landmarks.ModuleRoot(),
		wareStaging, // FUTURE: should probably use different warehouse for this, so it's easier to GC the shortlived objects
	}.Resolve
	pins, pinWs, err := funcs.ResolvePins(mod, viewCatalogTool, viewWarehousesTool, resolveTool)
	if err != nil {
		return errcat.Errorf(
			"stellar-resolve-imports-failed",
			"cannot resolve imports: %s", err)
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
	os.Setenv("REPEATR_MEMODIR", landmarks.MemoDir())
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
	//	for k, v := range exports {
	//		fmt.Fprintf(stderr, "  - %q: %v\n", k, v)
	//	}
	atl_exports := atlas.MustBuild(api.WareID_AtlasEntry)
	if err := refmt.NewMarshallerAtlased(json.EncodeOptions{Line: []byte("\n"), Indent: []byte("\t")}, stdout, atl_exports).Marshal(exports); err != nil {
		panic(err)
	}

	// If we have a saga name and we're not an anon module, save a "candidate" release!
	if sagaName == nil {
		return nil
	}
	modName, err := workspace.ResolveModuleName(landmarks)
	if err != nil {
		return err // FIXME detect this WAY earlier.
	}
	if modName == nil {
		return nil
	}
	if err := catalog.SaveCandidateRelease(landmarks.Workspace, *sagaName, *modName, exports, stderr); err != nil {
		return err
	}
	if err := catalog.SaveCandidateReplay(landmarks.Workspace, *sagaName, *modName, mod, stderr); err != nil {
		return err
	}
	return nil
}
