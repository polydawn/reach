package layout

import (
	"path/filepath"

	"go.polydawn.net/go-timeless-api"
)

// Landmarks holds all the major context-defining filesystem paths.
// The module root path, workspace root path, and everything that can be
// extrapolated from those is available on methods of this struct.
//
// If FindLandmarks started with a relative path, all landmarks will be
// relative; if it was absolute, they'll all be absolute.
//
// Note that no workspace configuration is loaded or parsed to determine
// any of these paths.
type Workspace struct {
	// Path of dir which contains the `.timeless` dir.
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

type Module struct {
	Workspace

	// Path of dir which contains `module.tl` file.
	moduleRoot string
}

func (lm Module) ModuleRoot() string {
	return lm.moduleRoot
}
func (lm Module) ModuleFile() string {
	return filepath.Join(lm.moduleRoot, "module.tl")
}

// NewModule constructs a new layout.Module -- but is not intended for public use.
// Use `layout.FindModule`, or `workspace.Workspace.GetModuleLayout` instead;
// these methods match the semantics that should be used and check the invariants.
func NewModule(ws Workspace, root string) Module {
	return Module{ws, root}
}
