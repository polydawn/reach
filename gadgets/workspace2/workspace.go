package workspace

import (
	"path/filepath"
)

type Workspace struct {
	// The workspace root path.  Just a regular path; might be relative (presumably to cwd) or might be absolute.
	// Doesn't bother to remember basisPath vs searchPath from creation; doesn't matter in the future.
	path string

	cfg *WorkspaceConfig

	moduleIndex map[string]string // map{modulePath:ModuleName} -- this is a cache index, freely rebuilt.  it's stored, but grown automatically when we encounter modules, and pruned equally silently.  globs over module names read this, and don't read the filesystem unless flags state to do so.  filesystem paths and globs automatically grow this.
}

type WorkspaceConfig struct {
	moduleNamingDirectives ModuleNamingDirectives
}

type ModuleNamingDirectives struct {
	workspaceName string // modules found under the workspace dir and affected by no other naming directives will have a name starting with this, and then accumulating path segments between the workspace root and module root as additional segments.

	pathModuleNameOverrides map[string]string // map{modulePath:ModuleName} -- this is user config: paths will be assigned these specific names.

	//pathModuleNamePartOverrides map[string]string // like the above, but it only changes how a single dir is mapped into a name segment. // okay, no, this is getting too complicated and too silly.  We want features that help shorten dir paths; not features that make them a total house of mirrors.

	dirModuleNamePrefixOverrides map[string]string // map{dirPath:modulePrefix} -- this is user config: modules found under this dir will have a name starting with the override and then accumulating path segments between there and the module root.  Overrides workspaceName's effect for anything it affects.

	moduleNameSubstitutions map[string]string // sed-like patterns for overriding module names.  Runs after any of the other overrides, and operates just on the parsed moduleNames (the path isn't part of consideration anymore at this phase).
}

// NewWorkspace returns a workspace object, assuming that the rootPath given is reasonable.
// Using Load or other functions later will error if the rootPath doesn't actually
// contain the configuration we expect a workspace to have.
//
// Prefer using FindWorkspace, which does much more of the work needed in common
// user stories around finding relevant a workspace.
func NewWorkspace(rootPath string) *Workspace {
	return &Workspace{
		path: filepath.Clean(rootPath),
		// that's it; everything else is loaded later.
	}
}

// Path returns the workspace's path.
//
// Typically this should not need to be used
// (other functions on the workspace object work with it on your behalf),
// but it is useful to print in logs and user-facing messages.
func (ws *Workspace) Path() string {
	return ws.path
}

// Load loads the workspace configuration (if it isn't already), or errors
// if the configuration is malformed.
//
// If Load has previously been called, calling it again is a no-op.
func (ws *Workspace) Load() error {
	// TODO: load WorkspaceConfig.
	return nil
}
