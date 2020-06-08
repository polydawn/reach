package workspace

type Module struct {
	// A module *is not defined* without a workspace.
	// The same module config file may be interpreted to produce a different Module object depending on its Workspace.
	ws *Workspace

	path string

	// The module name is a synthesized property;
	// it's not present in the module config along, and needs to be determined using info from the Workspace.
	//
	// FIXME there's a type for this.
	name string
}
