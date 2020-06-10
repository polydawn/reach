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
type ErrWorkspaceNotFound struct {
	SearchedExactly string // Either SearchedExactly or SearchedFrom will be set.
	SearchedFrom    string // Either SearchedExactly or SearchedFrom will be set.
	SearchedUpTo    string // SeachedUpTo is set if there was a basis path that limited the search.
}

func (e ErrWorkspaceNotFound) Error() string {
	if e.SearchedExactly != "" {
		return fmt.Sprintf("workspace not found (searched for one at %q but found no '%s' dir)", e.SearchedExactly, magicWorkspaceDirname)
	}
	if e.SearchedUpTo == "" {
		return fmt.Sprintf("workspace not found (searched starting at %q and all the way up to the root without finding a '%s' dir)", e.SearchedFrom, magicWorkspaceDirname)
	}
	return fmt.Sprintf("workspace not found (searched starting at %q and up to %q without finding a '%s' dir)", e.SearchedFrom, e.SearchedUpTo, magicWorkspaceDirname)
}

// ErrModuleNotFound is returned from functions such as FindModule
// which attempt to detect module file patterns.
type ErrModuleNotFound struct {
	SearchedExactly string // Either SearchedExactly or SearchedFrom will be set.
	SearchedFrom    string // Either SearchedExactly or SearchedFrom will be set.
	SearchedUpTo    string // SeachedUpTo is set if there was a basis path that limited the search.
}

func (e ErrModuleNotFound) Error() string {
	if e.SearchedExactly != "" {
		return fmt.Sprintf("module not found (searched for one at %q but found no '%s' file)", e.SearchedExactly, magicModuleFilename)
	}
	if e.SearchedUpTo == "" {
		return fmt.Sprintf("module not found (searched starting at %q and all the way up to the root without finding a '%s' file)", e.SearchedFrom, magicModuleFilename)
	}
	return fmt.Sprintf("module not found (searched starting at %q and up to %q without finding a '%s' file)", e.SearchedFrom, e.SearchedUpTo, magicModuleFilename)
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
