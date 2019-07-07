package workspace

import (
	"fmt"
	"path"
	"strings"

	"github.com/warpfork/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/reach/gadgets/layout"
)

/*
	Organizational note: Yes, there are several things called "Workspace".

	- `layout.Workspace` --> just the filesystem paths.

	- `workspace.Workspace` --> (this here) loads and understands config.
*/

const anonymousModuleName api.ModuleName = api.ModuleName("unnamed")

type Workspace struct {
	Layout layout.Workspace

	// Future: should have some config for which paths inside the workspace
	//  are considered to contain modules.
	//  (And for bonus points, any name transform patterns.)
	//  This should work with globs, because otherwise scaling will be irritating.
	//  e.g. `"*": ""` if every dir is a fully-qualified module name.
	//  e.g. `"modules/*": "foo.example.org/"` if every dir under "modules/"
	//   should be treated as a module name prefixed with "foo.example.org/".
	// Right now we're pretending everything is `"*": ""` plus `".": "unnamed"`.
}

// ResolveModuleName returns the public name of a module based on its path
// within the workspace filesystem.
//
// By default, the path within the workspace to the module will be its name;
// the workspace configuration can specify alternate mappings.
// (As a special case, if the module path is the exact same as the workspace
// root, then it will be mapped to the module name "unnamed" -- this may also
// be overriden by workspace config.)
//
// The moduleLayout argument must be for a path inside the workspace;
// otherwise a panic will be raised.
func (ws Workspace) ResolveModuleName(moduleLayout layout.Module) (api.ModuleName, error) {
	wsRoot := ws.Layout.WorkspaceRoot()
	modRoot := moduleLayout.ModuleRoot()

	// Module must be inside workspace: sanity check.
	if !strings.HasPrefix(modRoot, wsRoot) {
		panic("moduleRoot must be within workspaceRoot")
	}

	// If the module is in the workspace root, it's definitely anon.
	if len(modRoot) == len(wsRoot) {
		return anonymousModuleName, nil
	}

	// Future: check for config that remaps paths<->names.

	// Munge path into name.
	modName := api.ModuleName(modRoot[len(wsRoot)+1:])

	// Sanity check name.
	if err := modName.Validate(); err != nil {
		return "", errcat.ErrorDetailed(
			hitch.ErrUsage,
			fmt.Sprintf("%q is not a valid module name: %s", modName, err),
			map[string]string{
				"ref": string(modName),
			})
	}
	return modName, nil
}

// GetModuleLayout returns a layout.Module describing where we'd expect
// such a module to be located.
//
// The module name must be matched by some part of the workspace's
// module mapping config, or nil will be returned.
// (Note: this feature is NYI, so nil will currently never be returned.)
//
// Providing an invalid ModuleName will panic.
//
// This function does not check whether or not a module file is actually
// *present*; just describes where it is expected to be if present.
// Use `module.Load` to actually load the content.
func (ws Workspace) GetModuleLayout(modName api.ModuleName) *layout.Module {
	// Sanity check modName.
	funcs.MustValidate(modName)

	// Anonymous modules point to the workspace root
	if modName == anonymousModuleName {
		modLayout := layout.NewModule(
			ws.Layout,
			ws.Layout.WorkspaceRoot(),
		)
		return &modLayout
	}
	// Future: check for config that remaps paths<->names.

	// Assemble and return the module layout description struct.
	modLayout := layout.NewModule(
		ws.Layout,
		path.Join(ws.Layout.WorkspaceRoot(), string(modName)),
	)
	return &modLayout
}
