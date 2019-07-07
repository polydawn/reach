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
	"go.polydawn.net/reach/gadgets/catalog"
	hitchGadget "go.polydawn.net/reach/gadgets/catalog/hitch"
	"go.polydawn.net/reach/gadgets/ingest"
	"go.polydawn.net/reach/gadgets/layout"
	"go.polydawn.net/reach/gadgets/module"
	"go.polydawn.net/reach/gadgets/workspace"
)

func EvalModule(
	workspace workspace.Workspace, // needed to figure out if we have a moduleName.
	landmarks layout.Module, // needed in case of ingests with relative paths.
	sagaName *catalog.SagaName, // may have been provided as a flag.
	mod api.Module, // already helpfully loaded for us.
	stdout, stderr io.Writer,
) error {
	// Process the module DAG into a linear toposort of steps.
	//  Any impossible graphs inside the module will error out here
	//   (but we won't get to checking imports and ingests until later).
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

	// Configure defaults for warehousing.
	//  We'll always consider the workspace's local dirs as a data source;
	//  and we'll also use it as a place to store produced wares
	//   (both for intermediates, final exports, and ingests).
	//  The wareSourcing config may be accumulated along with others per formula;
	//   this is just the starting point minimum configuration.
	wareStaging := api.WareStaging{ByPackType: map[api.PackType]api.WarehouseLocation{"tar": landmarks.StagingWarehouseLoc()}}
	wareSourcing := api.WareSourcing{}
	wareSourcing.AppendByPackType("tar", landmarks.StagingWarehouseLoc())
	// Make the workspace's local warehouse dir if it doesn't exist.
	os.Mkdir(landmarks.StagingWarehousePath(), 0755)

	// Prepare catalog view tools.
	//  Definitely includes the workspace catalog;
	//  may also include a view of "candidates" data, if a sagaName arg is present.
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

	// Resolve all imports.
	//  This includes both viewing catalogs (cheap, fast),
	//  *and invoking ingest* (potentially costly).
	resolveTool := ingest.Config{
		landmarks.ModuleRoot(),
		wareStaging, // FUTURE: should probably use different warehouse for this, so it's easier to GC the shortlived objects
	}.Resolve
	pins, pinWs, err := funcs.ResolvePins(mod, viewCatalogTool, viewWarehousesTool, resolveTool)
	if err != nil {
		return errcat.Errorf(
			"reach-resolve-imports-failed",
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

	// Ensure memoization is enabled.
	//  Future: this is a bit of an odd reach-around way to configure this.
	//  PRs which propose more/better ways to enable and parameterize memoization would be extremely welcomed.
	os.Setenv("REPEATR_MEMODIR", workspace.Layout.MemoDir())
	os.Mkdir(workspace.Layout.MemoDir(), 0755) // Errors ignored.  Repeatr will emit warns, but work.

	// Begin the evaluation!
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

	// Print the results!
	//  This goes onto stdout as json, so it's parsible and pipeable.
	fmt.Fprintf(stderr, "module exports:\n")
	//	for k, v := range exports {
	//		fmt.Fprintf(stderr, "  - %q: %v\n", k, v)
	//	}
	atl_exports := atlas.MustBuild(api.WareID_AtlasEntry)
	if err := refmt.NewMarshallerAtlased(json.EncodeOptions{Line: []byte("\n"), Indent: []byte("\t")}, stdout, atl_exports).Marshal(exports); err != nil {
		panic(err)
	}

	// Save a "candidate" release!
	// (Unless there's no saga name, that must've been on purpose.)
	if sagaName == nil {
		return nil
	}
	modName, err := workspace.ResolveModuleName(landmarks)
	if err != nil {
		return err // FIXME detect this WAY earlier.
	}
	if err := catalog.SaveCandidateRelease(landmarks.Workspace, *sagaName, modName, exports, stderr); err != nil {
		return err
	}
	if err := catalog.SaveCandidateReplay(landmarks.Workspace, *sagaName, modName, mod, stderr); err != nil {
		return err
	}
	return nil
}
