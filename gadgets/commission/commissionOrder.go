package commission

import (
	"fmt"
	"sort"
	"strings"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/hitch"
	"go.polydawn.net/reach/gadgets/catalog"
	"go.polydawn.net/reach/gadgets/module"
	"go.polydawn.net/reach/gadgets/workspace"
)

/*
	CommissionOrder takes a list of ModuleName and returns a list of those
	ModuleNames and every additonal ModuleName which needs to be evaluated
	to satisfy all "candidate" imports, recursively.

	The order of ModuleName is a consistent topological sort,
	using the lexigraphical ordering of ModuleNames as a tiebreaker.
	Thus, evaluating each Module in the list, in order, results in a
	correct and complete evaluation of the whole set.
*/
func CommissionOrder(ws workspace.Workspace, wantList ...api.ModuleName) ([]api.ModuleName, error) {
	// Sort nodes by their name (this is our tiebreaker, in advance).
	nodesOrdered := make([]api.ModuleName, len(wantList))
	copy(nodesOrdered, wantList)
	sort.Sort(moduleNameByLex(nodesOrdered))
	// For each step: visit.  (This will recurse, and no-op itself internally as approrpriate for visited nodes.)
	visited := map[api.ModuleName]struct{}{}
	result := make([]api.ModuleName, 0, len(wantList))
	for _, node := range nodesOrdered {
		if err := orderModules_visit(ws, node, visited, []string{}, &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func orderModules_visit(
	ws workspace.Workspace,
	node api.ModuleName,
	visited map[api.ModuleName]struct{},
	backtrace []string,
	result *[]api.ModuleName,
) error {
	// First, check for cycles.  If this is in our current walk path already, bad.
	nBacktrace := len(backtrace)
	backtrace = append(backtrace, string(node))
	for i, backstep := range backtrace[:nBacktrace] {
		if backstep == string(node) {
			return fmt.Errorf("cycle found: %s", strings.Join(backtrace[i:nBacktrace], " -> "))
		}
	}

	// If we have visited this before (and not in a cycle), early out.
	if _, ok := visited[node]; ok {
		return nil
	}
	visited[node] = struct{}{}

	// Load module via the workspace.
	modLayout := ws.GetModuleLayout(node)
	mod, err := module.Load(*modLayout)
	if err != nil {
		return fmt.Errorf("error loading module: %s", err)
	}

	// Collect all imports.
	//  (This is its own function because it's recursive; it has to read submodules.)
	//  We just handle this as a flattened list; don't really care at which part
	//   of the module or submodule wants the import, just need to know the edges.
	imports := listImports(*mod)

	// Check that those actually point somewhere.
	//  - For catalogs: check that the catalog exists and has that version name.
	//  - For catalogs with "candidate" version: save to a new list: we'll recurse on these.
	//  - For parent refs: ignore it, the correctness of those is module's internal problem.
	//  - For ingests: ignore it, we assume it'll work itself out.
	candidateImports := []api.ModuleName(nil)
	for _, imp := range imports {
		switch imp2 := imp.(type) {
		case api.ImportRef_Catalog:
			switch imp2.ReleaseName {
			case "candidate":
				// Save these; this are the ones we care about to recurse.
				candidateImports = append(candidateImports, imp2.ModuleName)
			default:
				// Do a quick check that we'll be able to get this version.
				//  This will also be done when eval'ing the module,
				//   so strictly speaking we certainly don't *need* to here,
				//   but it's cheap enough to check now as well as later.
				catTree := catalog.Tree{ws.Layout.CatalogRoot()}
				modCat, err := catTree.LoadModuleLineage(imp2.ModuleName)
				if err != nil {
					return fmt.Errorf("unable to resolve import %q (wanted by module %q): %s",
						imp, node, err)
				}
				if _, err := hitch.LineagePluckReleaseItem(*modCat, imp2.ReleaseName, imp2.ItemName); err != nil {
					return fmt.Errorf("unable to resolve import %q (wanted by module %q): %s",
						imp, node, err)
				}
			}
		case api.ImportRef_Parent:
			// skip.
		case api.ImportRef_Ingest:
			// skip.
		default:
			panic("unreachable")
		}
	}

	// Sort the dependency nodes by name, then recurse.
	//  This sort is necessary for deterministic order of unrelated nodes.
	sort.Sort(moduleNameByLex(candidateImports))
	fmt.Printf("for module %q, discovered recursive links to: %v\n", node, candidateImports)
	for _, imp := range candidateImports {
		if err := orderModules_visit(ws, imp, visited, backtrace, result); err != nil {
			return err
		}
	}

	// Done: put this node in the results.
	//  It's important that we append ourselves *last*: toposort
	//   means all the things we depend on must be above us.
	*result = append(*result, node)
	return nil
}

func listImports(m api.Module) (refs []api.ImportRef) {
	// Add everything in this module to the list.
	for _, v := range m.Imports {
		refs = append(refs, v)
	}
	// Traverse steps, and recurse to accumulate across any submodules.
	for _, step := range m.Steps {
		switch x := step.(type) {
		case api.Operation:
			// pass.  hakuna matata; operations only have local references to their module's imports.
		case api.Module:
			// recurse, and contextualize all refs from the deeper module(s).
			refs = append(refs, listImports(x)...)
		}
	}
	return
}

type moduleNameByLex []api.ModuleName

func (a moduleNameByLex) Len() int           { return len(a) }
func (a moduleNameByLex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a moduleNameByLex) Less(i, j int) bool { return a[i] < a[j] }
