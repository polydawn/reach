package workspace

import (
	"os"
	"path/filepath"
)

// FindWorkspace finds a workspace by looking at each segment of the searchPath,
// checking if it has the directories that indicate a workspace root,
// and either returning the first one found,
// or otherwise continuing to search further up the path until one is found,
// or the path runs out of segments.
// When a workspace is found, any remaining path segments will be returned also.
//
// The indicator for a workspace root is the presence of a '.timeless' directory;
// the directory containing this marker directory is consider the workspace root.
// (There will likely be config and additional files found within the '.timeless'
// directory, but this function does not check for them nor load them;
// use the 'Load' function on the returned Workspace handle for that.)
//
// The searchPath may be absolute or relative.
// If searchPath is relative, it will not be absolutized;
// a relative searchPath may effectively be used to limit the search area.
// If basisPath is provided, it is conjoined to the front of the path before opening it,
// but is not included in the search.
//
// Even if several segments of searchPath don't exist, the search won't return early;
// it just quietly continues searching up segment after segment.
// A filesystem "not found" error will never be returned;
// if we consume every segment of the searchPath without finding a workspace,
// we return ErrWorkspaceNotFound.
//
// Workspaces may be nested.  When a workspace is found,
// the remainingSearchPath value returned may be used in a subsequent FindWorkspace call
// to find parents of the returned workspace.
//
// The most typical form of usage in the reach application overall
// is a blank basisPath with an absolutized searchPath that is the cwd;
// this results in the semantic behavior of "search all directories
// up from the cwd until you find a workspace".
//
func FindWorkspace(basisPath, searchPath string) (ws *Workspace, remainingSearchPath string, err error) {
	// Our search loops over searchPath, popping a path segment off at the end of every round.
	//  Keep the given searchPath in hand; we might need it for an error report.
	searchAt := searchPath
	for {
		// Assume the search path exists and is a dir (we'll get a reasonable error anyway if it's not);
		//  join that path with our search target and try to open it.
		f, err := os.Open(filepath.Join(basisPath, searchAt, magicWorkspaceDirname))
		f.Close()
		if err == nil { // no error?  Found it!
			return NewWorkspace(filepath.Join(basisPath, searchAt)), filepath.Dir(searchAt), nil
		}
		if os.IsNotExist(err) { // no such thing?  oh well.  pop a segment and keep looking.
			searchAt = filepath.Dir(searchAt)
			// If popping a searchAt segment got us down to nothing,
			//  and we didn't find anything here either,
			//   that's it: return NotFound.
			if searchAt == "/" || searchAt == "." {
				return nil, "", &ErrWorkspaceNotFound{"", filepath.Join(basisPath, searchPath), basisPath}
			}
			// ... otherwise: continue, with popped searchAt.
			continue
		}
		// You're still here?  That means there's an error, but of some unpleasant kind.
		//  Whatever this error is, our search has blind spots: error out.
		return nil, searchAt, &ErrSearchingFilesystem{"workspace", err}
	}
}

// FindModule looks for a module within the searchPath.
// It's very similar to how FindWorkspace works, except the indicator it's looking
// for is a directory that contains a 'module.tl' file.
//
// Though modules can be nested, this function doesn't return remainingSearchPath.
// If one wants to regard multiple modules at once, one will usually want the FindModules function instead,
// and as a result there's rarely any need to regard remainingSearchPath.
//
// The basisPath doesn't have to match the basisPath used when finding the Workspace,
// but if the join of basisPath and searchPath ever leaves the workspace's root,
// then the search will halt and return ErrModuleNotFound.
func FindModule(ws *Workspace, basisPath, searchPath string) (*Module, error) {
	// Our search loops over searchPath, popping a path segment off at the end of every round.
	//  Keep the given searchPath in hand; we might need it for an error report.
	searchAt := searchPath
	for {
		mod, err := ExpectModule(ws, basisPath, searchAt)
		switch err.(type) {
		case nil: // no error?  found it!
			return mod, nil
		case *ErrModuleNotFound: // not found?  oh well.  pop a segment and keep looking.
			searchAt = filepath.Dir(searchAt)
			// If popping a searchAt segment got us down to nothing,
			//  and we didn't find anything here either,
			//   that's it: return NotFound.
			if searchAt == "/" || searchAt == "." {
				return nil, &ErrModuleNotFound{"", filepath.Join(basisPath, searchPath), basisPath}
			}
			// ... otherwise: continue, with popped searchAt.
		default: // other error?  alarming; keep raising it.
			return nil, err
		}
	}
}

// ExpectModule expects to find a module root at exactly the given tryPath.
// It will not search upwards in parent directories if it doesn't find one there,
// like FindModule would; instead it returns ErrModuleNotFound immediately.
//
// If the join of basisPath and tryPath is not within the Workspace root,
// ErrNotInWorkspace will be returned, regardless of whether module indicators exist.
// (Implicitly, the basisPath can be longer than the basisPath used to find the workspace,
// or it can be shorter, and a longer tryPath makes up for it;
// but if the basisPath doesn't share a prefix with a prefix the basisPath used to find the workspace,
// then there's no chance that the search will be considered to fall inside the workspace root.)
func ExpectModule(ws *Workspace, basisPath, tryPath string) (*Module, error) {
	fullTryPath := filepath.Join(basisPath, tryPath)
	if !filepath.HasPrefix(fullTryPath, ws.path) {
		return nil, &ErrNotInWorkspace{}
	}
	match, err := PathContainsModuleIndicators(fullTryPath)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, &ErrModuleNotFound{fullTryPath, "", ""}
	}
	return &Module{ws, fullTryPath, "TODO: moduleName computation"}, nil // ... it's disturbing that this cast works.  We should replace that type with a `type ModuleName{ x string }` wrapper type.  I grow tired of Golang's type system... lazyness.  It pushes a lot of boilerplate on to me; if I want any sanity enforcement at all, it rapidly gets almost as bad as java.
}

// PathContainsModuleIndiactors simply checks if the given path contains
// the 'module.tl' file that indicates a module root.
//
// This is not very useful on its own (FindModule or ExpectModule are more likely useful),
// but can be used to make more helpful guidance messages to a user.
//
// An error is only returned if it's ErrSearchingFilesystem
// ("not found" is simply represented by a return of 'false').
func PathContainsModuleIndicators(pth string) (bool, error) {
	f, err := os.Open(filepath.Join(pth, magicModuleFilename))
	f.Close()
	if err == nil { // no error?  Found it!
		return true, nil
	}
	if os.IsNotExist(err) { // no such thing?  oh well.
		return false, nil
	}
	// You're still here?  That means there's an error, but of some unpleasant kind.
	//  Whatever this error is, our search has blind spots: error out.
	return false, &ErrSearchingFilesystem{"module", err}
}

func FindModules(ws *Workspace, basisPath, visitFn func(*Module) error, tryPaths ...string) error {
	panic("nyi")
	// do we want some sort of WorkspaceUpdateCurrier param to this, too?  I think when we do this search, we usually want that.
	// but maybe that's better done as wee util that you put in your visitor func.
}

// FUTURE: SearchModules func, which takes search patterns instead of just a flat set of paths to try.
