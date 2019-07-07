package waresApp

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"github.com/polydawn/refmt/obj/atlas"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/go-timeless-api/rio"
	rioclient "go.polydawn.net/go-timeless-api/rio/client/exec"
	"go.polydawn.net/reach/gadgets/catalog"
	hitchGadget "go.polydawn.net/reach/gadgets/catalog/hitch"
	"go.polydawn.net/reach/gadgets/workspace"
)

func ListCandidates(ws workspace.Workspace, moduleName api.ModuleName, sagaName catalog.SagaName, itemName *api.ItemName, stdout, stderr io.Writer) error {
	tree := catalog.Tree{filepath.Join(ws.Layout.WorkspaceRoot(), ".timeless/candidates/", sagaName.String())}
	lineage, err := tree.LoadModuleLineage(moduleName)
	if err != nil {
		return err
	}

	if itemName == nil {
		release, err := hitch.LineagePluckReleaseByName(*lineage, "candidate")
		if err != nil {
			return err
		}
		atl_exports := atlas.MustBuild(api.WareID_AtlasEntry)
		if err := refmt.NewMarshallerAtlased(
			json.EncodeOptions{Line: []byte("\n"), Indent: []byte("\t")},
			stdout,
			atl_exports,
		).Marshal(release.Items); err != nil {
			panic(err)
		}
	} else {
		wareID, err := hitch.LineagePluckReleaseItem(*lineage, "candidate", *itemName)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "%s\n", wareID)
	}

	return nil
}

func ListReleases(ws workspace.Workspace, moduleName api.ModuleName, releaseName *api.ReleaseName, itemName *api.ItemName, stdout, stderr io.Writer) error {
	viewLineageTool, _ := hitchGadget.ViewTools([]catalog.Tree{
		// refactor note: we used to stack several catalog dirs here, but have backtracked from allowing that.
		// so it's possible there's a layer of abstraction here that should be removed outright; have not fully reviewed.
		{ws.Layout.CatalogRoot()},
	}...)

	lineage, err := viewLineageTool(context.TODO(), moduleName)
	if err != nil {
		return err
	}
	if releaseName != nil && itemName != nil {
		wareID, err := hitch.LineagePluckReleaseItem(*lineage, *releaseName, *itemName)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "%s\n", wareID)
	} else if releaseName != nil {
		release, err := hitch.LineagePluckReleaseByName(*lineage, *releaseName)
		if err != nil {
			return err
		}
		atl_exports := atlas.MustBuild(api.WareID_AtlasEntry)
		if err := refmt.NewMarshallerAtlased(
			json.EncodeOptions{Line: []byte("\n"), Indent: []byte("\t")},
			stdout,
			atl_exports,
		).Marshal(release.Items); err != nil {
			panic(err)
		}
	} else {
		output := make(map[api.ReleaseName]map[api.ItemName]api.WareID)
		for _, release := range lineage.Releases {
			for item, wareId := range release.Items {
				if output[release.Name] == nil {
					output[release.Name] = make(map[api.ItemName]api.WareID)
				}
				output[release.Name][item] = wareId
			}
		}
		if len(output) > 0 {
			atl_exports := atlas.MustBuild(api.WareID_AtlasEntry)
			if err := refmt.NewMarshallerAtlased(
				json.EncodeOptions{Line: []byte("\n"), Indent: []byte("\t")},
				stdout,
				atl_exports,
			).Marshal(output); err != nil {
				panic(err)
			}
		}
	}

	return nil
}
func UnpackCandidate(ctx context.Context, ws workspace.Workspace, sagaName catalog.SagaName, moduleName api.ModuleName, itemName api.ItemName, path string, stdout, stderr io.Writer) error {
	tree := catalog.Tree{filepath.Join(ws.Layout.WorkspaceRoot(), ".timeless/candidates/", sagaName.String())}
	lineage, err := tree.LoadModuleLineage(moduleName)
	if err != nil {
		return err
	}
	wareID, err := hitch.LineagePluckReleaseItem(*lineage, "candidate", itemName)
	if err != nil {
		return err
	}

	wareSourcing := api.WareSourcing{}
	wareSourcing.AppendByPackType("tar", ws.Layout.StagingWarehouseLoc())
	_, viewWarehouseTool := hitchGadget.ViewTools([]catalog.Tree{
		// refactor note: we used to stack several catalog dirs here, but have backtracked from allowing that.
		// so it's possible there's a layer of abstraction here that should be removed outright; have not fully reviewed.
		{ws.Layout.CatalogRoot()},
		{filepath.Join(ws.Layout.WorkspaceRoot(), ".timeless/candidates/", sagaName.String())},
	}...)
	// viewWarehouseTool = hitchGadget.WithCandidates(
	// 	viewWarehouseTool,
	// 	catalog.Tree{filepath.Join(ws.Layout.WorkspaceRoot(), ".timeless/candidates/", sagaName.String())},
	// )
	warehouse, err := viewWarehouseTool(ctx, moduleName)
	wareSourcing.Append(*warehouse)
	wareSourcing = wareSourcing.PivotToModuleWare(*wareID, moduleName)
	warehouseLocations := []api.WarehouseLocation{
		ws.Layout.StagingWarehouseLoc(),
	}
	return UnpackWareContents(ctx, ws, warehouseLocations, *wareID, path, stdout, stderr)
}

func UnpackRelease(ctx context.Context, ws workspace.Workspace, moduleName api.ModuleName, releaseName api.ReleaseName, itemName api.ItemName, path string, stdout, stderr io.Writer) error {
	viewLineageTool, _ := hitchGadget.ViewTools([]catalog.Tree{
		// refactor note: we used to stack several catalog dirs here, but have backtracked from allowing that.
		// so it's possible there's a layer of abstraction here that should be removed outright; have not fully reviewed.
		{ws.Layout.CatalogRoot()},
	}...)

	lineage, err := viewLineageTool(ctx, moduleName)
	if err != nil {
		return err
	}
	wareID, err := hitch.LineagePluckReleaseItem(*lineage, releaseName, itemName)
	if err != nil {
		return err
	}
	wareSourcing := api.WareSourcing{}
	wareSourcing.AppendByPackType("tar", ws.Layout.StagingWarehouseLoc())
	_, viewWarehouseTool := hitchGadget.ViewTools([]catalog.Tree{
		// refactor note: we used to stack several catalog dirs here, but have backtracked from allowing that.
		// so it's possible there's a layer of abstraction here that should be removed outright; have not fully reviewed.
		{ws.Layout.CatalogRoot()},
	}...)
	warehouse, err := viewWarehouseTool(ctx, moduleName)
	wareSourcing.Append(*warehouse)
	wareSourcing = wareSourcing.PivotToModuleWare(*wareID, moduleName)

	return UnpackWareContents(ctx, ws, wareSourcing.ByWare[*wareID], *wareID, path, stdout, stderr)
}

func UnpackWareID(ctx context.Context, ws workspace.Workspace, wareId api.WareID, path string, stdout, stderr io.Writer) error {
	wareSourcing := api.WareSourcing{}
	wareSourcing.AppendByPackType("tar", ws.Layout.StagingWarehouseLoc())
	wareSourcing = wareSourcing.PivotToModuleWare(wareId, "")
	return UnpackWareContents(ctx, ws, wareSourcing.ByWare[wareId], wareId, path, stdout, stderr)
}

func UnpackWareContents(ctx context.Context, ws workspace.Workspace, wareLocations []api.WarehouseLocation, wareId api.WareID, path string, stdout, stderr io.Writer) error {
	unpackedId, err := rioclient.UnpackFunc(ctx,
		wareId,
		path,
		api.FilesetUnpackFilter_LowPriv,
		rio.Placement_Direct,
		wareLocations,
		rio.Monitor{},
	)
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Unpacked WareID: %s\n", unpackedId)
	return nil
}
