package layout

import (
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
)

// layout.Workspace holds the filesystem paths defining a workspace.
//
// Note that the `workspace.Workspace` loads and understands config --
// so most complex logic goes there and uses that type instead.
// Use `layout.Workspace` when you *only* want to consider the paths
// as a landmark and not yet actually parse the workspace config.
//
// layout.Workspace is produced by layout.FindWorkspace.
// If FindWorkspace started with a relative path, all paths will be
// relative; if it was absolute, they'll all be absolute.
type Workspace struct {
	// Path of dir which contains the `.timeless` dir.
	// The string does *not* include the `.timeless` suffix.
	workspaceRoot string
}

func (lm Workspace) WorkspaceRoot() string {
	return lm.workspaceRoot
}
func (lm Workspace) WorkspaceConfigFile() string {
	return filepath.Join(lm.workspaceRoot, ".timeless", "workspace.tl")
}
func (lm Workspace) CatalogRoot() string {
	// review: would we get better log messages if we resolved any symlinks first?
	return filepath.Join(lm.workspaceRoot, ".timeless", "catalog")
}
func (lm Workspace) MemoDir() string {
	// review: would we get better log messages if we resolved any symlinks first?
	return filepath.Join(lm.workspaceRoot, ".timeless", "memo")
}
func (lm Workspace) StagingWarehousePath() string {
	// review: would we get better log messages if we resolved any symlinks first?
	return filepath.Join(lm.workspaceRoot, ".timeless", "warehouse")
}
func (lm Workspace) StagingWarehouseLoc() api.WarehouseLocation {
	// review: would we get better log messages if we resolved any symlinks first?
	return api.WarehouseLocation("ca+file://" + filepath.Join(lm.workspaceRoot, ".timeless", "warehouse"))
}

// layout.Module holds the filesystem paths defining a module.
//
// Note that `module.Load()` can be used to convert this into a fully-parsed
// and usable `api.Module` type.
//
// Note that a ModuleName can not be inferred from paths alone, and so is
// also not present in this type: use `workspace.Workspace.ResolveModuleName()`
// to determine a ModuleName.
type Module struct {
	Workspace

	// Path of dir which contains `module.tl` file.
	// Is fully qualified (or at least as much as the workspace path was).
	moduleRoot string
}

// ModuleRoot returns the filesystem path to the root directory of the module.
// This path is already prefixed with the workspace root path (and is absolute
// if the workspace path was absolute; and relative if it was relative).
func (lm Module) ModuleRoot() string {
	return lm.moduleRoot
}

// ModuleFile returns the filesystem path to the module file.
func (lm Module) ModuleFile() string {
	return filepath.Join(lm.moduleRoot, "module.tl")
}

// NewModule constructs a new layout.Module -- but is not intended for public use.
// Use `layout.FindModule`, or `workspace.Workspace.GetModuleLayout` instead;
// this method does *not* check invariants and is not safe to use directly.
func NewModule(ws Workspace, root string) Module {
	return Module{ws, root}
}
