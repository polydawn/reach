package workspace

import (
	"fmt"
)

// ErrSearchingFilesystem is the error returned when searching the filesystem
// encounters some error (most typically, this is filesystem permissions errors).
type ErrSearchingFilesystem struct {
	For   string // either "workspace" or "module"
	Cause error
}

func (e ErrSearchingFilesystem) Error() string {
	return fmt.Sprintf("error searching for %s: %s", e.For, e.Cause)
}
func (e ErrSearchingFilesystem) Unwrap() error {
	return e.Cause
}

// ErrWorkspaceNotFound is returned from functions such as FindWorkspace
// which attempt to detect workspace directory patterns.
// It doesn't include any reasons; the parameters to whatever function
// returned this error are probably the relevant information to include
// in any larger report.
type ErrWorkspaceNotFound struct{}

func (e ErrWorkspaceNotFound) Error() string {
	return "workspace not found"
}

// ErrModuleNotFound is returned from functions such as FindModule
// which attempt to detect module file patterns.
// It doesn't include any reasons; the parameters to whatever function
// returned this error are probably the relevant information to include
// in any larger report.
type ErrModuleNotFound struct{}

func (e ErrModuleNotFound) Error() string {
	return "module not found"
}

// ErrNotInWorkspace is returned from functions such as FindModule
// when their search would be about to leave the Workspace's directory.
// If this happens as the search has widened, the SearchPath may be
// different than the search path you input to the function that errored.
type ErrNotInWorkspace struct {
	WorkspaceRoot string
	SearchPath    string
}

func (e ErrNotInWorkspace) Error() string {
	return fmt.Sprintf("searching a path not inside a workspace (%q is not contained in %q)", e.SearchPath, e.WorkspaceRoot)
}
