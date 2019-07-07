package waresApp

import (
	"fmt"
	"io"
	"path/filepath"
	
	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"github.com/polydawn/refmt/obj/atlas"
	
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/reach/gadgets/catalog"
	"go.polydawn.net/reach/gadgets/layout"
	"go.polydawn.net/reach/gadgets/workspace"
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
