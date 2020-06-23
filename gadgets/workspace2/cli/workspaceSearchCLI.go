package workspacecli

import (
	"fmt"

	"github.com/urfave/cli"

	workspace "go.polydawn.net/reach/gadgets/workspace2"
)

// this file doesn't need a package for any real technical reason, does it?
// possibly still helpful for clarity, but, can be debated.
//
// i'm actually getting quickly upset with not having these tests run from the package.

// The "search-workspace" command is a plumbing command that searches for workspaces and modules,
// and then prints basic information about their path (and names, if possible) to stdout as plain text.
//
// As a plumbing command, users won't usually need to run this to get things done
// (other more user-facing commands will implicitly use these same operations, internally).
// We ship it in case it's useful in debugging, or understanding what Reach is thinking.
var Cmd = &cli.Command{
	Name:     "search-workspace",
	Category: "plumbing",
	Action: func(args *cli.Context) error {
		wd := args.Path("wd")

		ws, remainingPath, err := workspace.FindWorkspace("", wd)
		if err != nil {
			fmt.Fprintf(args.App.Writer, "workspace:        -none- (%s)\n", err)
			return nil
		}
		fmt.Fprintf(args.App.Writer, "workspace:        %s\n", ws.Path())

		pws, _, err := workspace.FindWorkspace("", remainingPath)
		if err != nil {
			fmt.Fprintf(args.App.Writer, "parent workspace: -none- (%s)\n", err)
		} else {
			fmt.Fprintf(args.App.Writer, "parent workspace: %s\n", pws.Path())
		}

		mod, err := workspace.FindModule(ws, "", wd)
		if err != nil {
			fmt.Fprintf(args.App.Writer, "module root:      -none- (%s)\n", err)
		} else {
			fmt.Fprintf(args.App.Writer, "module root:      %s\n", mod.Path())
			fmt.Fprintf(args.App.Writer, "module name:      %s\n", mod.Name())
		}

		// JSON mode?  Not yet.
		// Commands that are quick to run, return one (or a small fairly fixed arity) string, have no arduous escaping needs, and are things we invision as being useable in shell scripting (without needing jq): plaintext works well for these.
		// (That sounds like a long list of conditions, but it's nonetheless fairly common.)
		//
		// ... of course, you got ambiguity and "escaping needs" almost immediately: look at how to distinguish errors, up there.
		// So, we do want to add JSON mode for this command's output, just like everything else.
		// (It would be nice to improve our tooling for doing this.  IPLD Schemas and codegen seem like they could help a great deal, "soon".
		// We should also move to having error structures be described in an language-independent schema... but this truly becomes a strict depends-on more tooling support.)
		//
		// We should probably make a struct of WorkspaceTypicalInfo which contains all this stuff.  It seems likely that's a reasonable api to export and be using across the program.

		return nil
	},
}
