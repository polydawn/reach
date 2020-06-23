package workspace

import (
	"path"
	"strings"

	api "go.polydawn.net/go-timeless-api"
)

// computes the module name, loading the workspace config (if that hasn't happened already).
// The 'modulePath' arg should be the path *within* the workspace (not an absolute path;
// no leading nor trailing dots nor slashes).
//
// Can error if the workspace config is invalid or loading it fails;
// can also error if the computed name is invalid (nasty characters, etc).
func computeModuleName(ws *Workspace, modulePath string) (api.ModuleName, error) {
	if err := ws.Load(); err != nil {
		return "", err
	}
	nd := ws.cfg.moduleNamingDirectives
	var mn api.ModuleName
	prefix := modulePath
	suffix := ""

	// Try direct overrides.
	if override, exists := nd.pathModuleNameOverrides[modulePath]; exists {
		mn = api.ModuleName(override)
		goto validate
	}

	// Try to find prefix overrides.
	for prefix != "." {
		if override, exists := nd.dirModuleNamePrefixOverrides[prefix]; exists {
			mn = api.ModuleName(path.Join(override, suffix))
			goto substitutions
		}
		suffix = path.Join(path.Base(prefix), suffix)
		prefix = path.Dir(prefix)
	}

	// No root overrides applied?  Okay: default to the workspaceName as a prefix.
	mn = api.ModuleName(path.Join(nd.workspaceName, modulePath))

substitutions:
	// Apply substitutions.
	//  This is pretty brute-force; but we don't expect there to be many of these (even a dozen would be eyebrow-raising).
	//  Behavior in the face of overlapping rules is currently **not defined**.
	//   (We could make this follow declaration order if we had order-preserving maps (which IPLD will give us).
	//    Sorting by length could also be a reasonably deterministic and probably-practically-viable rule.)
	for pattern, substitution := range nd.moduleNameSubstitutions {
		mn = api.ModuleName(strings.Replace(string(mn), pattern, substitution, 1))
	}

validate:
	if err := mn.Validate(); err != nil {
		return "", err
	}
	return mn, nil
}

// ModuleNamingDirectives are part of workspace configuration that is used to determine how paths are mapped into module names.
// Everything in this structure is mappable to serial configuration files.
type ModuleNamingDirectives struct {
	// TODO: the claim about "mappable to serial configuration files" needs to, well, happen.

	// A workspace always has a "name" that's a default prefix for all other naming.
	// Modules found under the workspace dir and affected by no other naming directives will have a name starting with this, and
	//  then accumulating path segments between the workspace root and module root as additional segments.
	workspaceName string

	// Workspace configuration can override the prefix for specific directories.
	// Modules found under such a dir will have a name starting with the override (instead of the workspaceName and any other accumulated path segments), and
	//  then accumulate path segments between that dir and the module root.
	//  In other words, this overrides workspaceName's effect for anything it affects.
	//
	// Only one prefix override can apply to a path:
	// the first prefix that matches as the path is searched from longest to shortest prefix will be the one used.
	//
	// The type is stringy, but is effectively `map{dirPath:modulePrefix}`.
	dirModuleNamePrefixOverrides map[string]string

	// Workspace configuration can override the module name for a specific directory.
	// This is somewhat similar to dirModuleNamePrefixOverrides except it only applies to *one* directory
	// (and not to any directories under it).
	//
	// These overrides are final: no other substitution rules apply if one of these overrides applied.
	//
	// The type is stringy, but is effectively `map{modulePath:ModuleName}`.
	pathModuleNameOverrides map[string]string

	// Workspace configuration can specify substitutions that apply to module paths.
	// These substitutions apply after all other rules, so they can stack with other overrides
	// (except for pathModuleNameOverrides, which are final).
	//
	// These substitutions are applied purely textually (they don't regard path segments):
	// `"foo":"bar"` applied to `example.org/foo/thing` will become `example.org/bar/thing`;
	// `"ab/cd":"ab/de"` applied to `example.org/zab/cdef` will become `example.org/zab/deef`.
	//
	// One interesting example usage is to map a path segment into nothing at all:
	// this can be done with a configuration like `"/foobar/":"/"`,
	// and can be used to create more directories to organize code in a workspace than are actually represented in module names.
	moduleNameSubstitutions map[string]string
}
