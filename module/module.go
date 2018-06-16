package module

import (
	"os"
	"path/filepath"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/stellar/layout"
)

func LoadByPath(ti layout.TreeInfo, pth string) (mod *api.Module, err error) {
	f, err := os.Open(filepath.Join(ti.Root, "module.tl"))
	if err != nil {
		return
	}
	err = refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, f, api.Atlas_Module).Unmarshal(&mod)
	return
}
