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
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/reach/gadgets/catalog"
	hitchGadget "go.polydawn.net/reach/gadgets/catalog/hitch"
	"go.polydawn.net/reach/gadgets/layout"
	"go.polydawn.net/reach/gadgets/workspace"
	rioclient "go.polydawn.net/go-timeless-api/rio/client/exec"
)

func ListCandidates(ws workspace.Workspace, layoutModule layout.Module, sagaName catalog.SagaName, itemName *api.ItemName, stdout, stderr io.Writer) error {
	tree := catalog.Tree{filepath.Join(ws.Layout.WorkspaceRoot(), ".timeless/candidates/", sagaName.String())}
	moduleName, err := ws.ResolveModuleName(layoutModule)
	if err != nil {
		return err
	}
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

func UnpackWareContents(context context.Context, ws workspace.Workspace, wareId api.WareID, path string, stdout,stderr io.Writer) error {
	unpackedId, err := rioclient.UnpackFunc(context,
		wareId,
		path,
		api.FilesetUnpackFilter_LowPriv,
		rio.Placement_Direct,
		[]api.WarehouseLocation {
			ws.Layout.StagingWarehouseLoc(),
		},
		rio.Monitor{},
	)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Unpacked WareID: %s\n", unpackedId)
	return nil
}
