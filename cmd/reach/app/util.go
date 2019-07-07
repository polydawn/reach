package reach

import (
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/reach/gadgets/layout"
	"go.polydawn.net/reach/gadgets/workspace"
	"path/filepath"
)

func ModuleNameOrPath(ws workspace.Workspace, modNameOrPath, curDir string) (*api.ModuleName, error) {
	if layout.IsModuleName(modNameOrPath) {
		modName := api.ModuleName(modNameOrPath)
		if err := modName.Validate(); err != nil {
			return nil, err
		}
		return &modName, nil
	} else {
		modulePath := filepath.Join(curDir, modNameOrPath)
		module, err := layout.ExpectModule(ws.Layout, modulePath)
		if err != nil {
			return nil, err
		}
		modName, err := ws.ResolveModuleName(*module)
		if err != nil {
			return nil, err
		}

		return &modName, nil
	}
}
