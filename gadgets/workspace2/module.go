package workspace

import (
	api "go.polydawn.net/go-timeless-api"
)

type Module struct {
	// A module *is not defined* without a workspace.
	// The same module config file may be interpreted to produce a different Module object depending on its Workspace.
	ws *Workspace

	path string

	// The module name is a synthesized property;
	// it's not present in the module config along, and needs to be determined using info from the Workspace.
	name api.ModuleName
}

// Path returns the module's path.
//
// Typically this should not need to be used
// (other functions on the module object work with it on your behalf),
// but it is useful to print in logs and user-facing messages.
func (mod *Module) Path() string {
	return mod.path
}

func (mod *Module) Name() api.ModuleName {
	return mod.name
}
