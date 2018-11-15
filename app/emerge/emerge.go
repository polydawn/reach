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
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/gadgets/catalog"
	hitchGadget "go.polydawn.net/stellar/gadgets/catalog/hitch"
	"go.polydawn.net/stellar/gadgets/ingest"
	"go.polydawn.net/stellar/gadgets/layout"
	"go.polydawn.net/stellar/gadgets/module"
)

func EvalModule(landmarks layout.Landmarks, sagaName *catalog.SagaName, mod api.Module, stdout, stderr io.Writer) error {
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
	viewCatalogTool, viewWarehousesTool := hitchGadget.ViewTools([]catalog.Tree{
		{landmarks.ModuleCatalogRoot},
		{filepath.Join(landmarks.WorkspaceRoot, ".timeless/catalogs/upstream")}, // TODO fix hardcoded "upstream" param
	}...)
	if sagaName != nil {
		viewCatalogTool = hitchGadget.WithCandidates(
			viewCatalogTool,
			catalog.Tree{filepath.Join(landmarks.WorkspaceRoot, ".timeless/candidates/", sagaName.String())},
		)
	}
	resolveTool := ingest.Config{
		landmarks.ModuleRoot,
		wareStaging, // FUTURE: should probably use different warehouse for this, so it's easier to GC the shortlived objects
	}.Resolve
	pins, pinWs, err := funcs.ResolvePins(mod, viewCatalogTool, viewWarehousesTool, resolveTool)
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
	//	for k, v := range exports {
	//		fmt.Fprintf(stderr, "  - %q: %v\n", k, v)
	//	}
	atl_exports := atlas.MustBuild(api.WareID_AtlasEntry)
	if err := refmt.NewMarshallerAtlased(json.EncodeOptions{Line: []byte("\n"), Indent: []byte("\t")}, stdout, atl_exports).Marshal(exports); err != nil {
		panic(err)
	}

	// If we have a saga name and we're in a workspace, save results!
	//  These results will become available as a "candidate" release.
	//  (It doesn't make sense to do this outside of a workspace, because no
	//   one would be able to use it; nor could we guess our own module name.)
	if sagaName != nil && landmarks.WorkspaceRoot != "" {
		modName := api.ModuleName(landmarks.ModulePathInsideWorkspace)
		if err := modName.Validate(); err != nil {
			return errcat.ErrorDetailed(
				hitch.ErrUsage,
				fmt.Sprintf("module %q: not a valid module name: %s", modName, err),
				map[string]string{
					"ref": string(modName),
				})
		}
		if err := catalog.SaveCandidateRelease(landmarks, *sagaName, modName, exports, stderr); err != nil {
			return err
		}
		if err := catalog.SaveCandidateReplay(landmarks, *sagaName, modName, mod, stderr); err != nil {
			return err
		}
	}
	return nil
}
