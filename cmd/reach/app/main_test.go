package reach_test

import (
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/reach/examples/testutil"
)

func TestNoargsHelptext(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("reach")
	Wish(t, exitCode, ShouldEqual, 0)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		NAME:
		   reach

		USAGE:
		   Reach is a multipurpose tool for driving and managing Timeless Stack projects.
		   Major functions of `+"`reach`"+` include:

		     - evaluating modules, which run pipelines of repeatr operations;
		     - staging release candidates of the results;
		     - commissioning entire generations of builds from many modules;
		     - and publishing and managing release catalogs for distribution.

		   See https://repeatr.io/ for more complete documention!

		COMMANDS:
		     emerge   evaluate a pipeline, logging intermediate results and reporting final exports
		     ci       given a module with one ingest using git, build it once, then build it again each time the git repo updates
		     catalog  catalog subcommands help maintain the release catalog info tree
		     wares    look up wares by release or candidate
		     help, h  Shows a list of commands or help for one command

		GLOBAL OPTIONS:
		   --help, -h     show help
		   --version, -v  print the version
	`))
}

func TestWrongCommandHelptext(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("reach", "not a command")
	Wish(t, exitCode, ShouldEqual, 1)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		reach: incorrect usage: 'not a command' is not a reach subcommand
	`))

	t.Run("also when asking for help", func(t *testing.T) {
		exitCode, stdout, stderr := RunIntoBuffer("reach", "not a command", "-h")
		Wish(t, exitCode, ShouldEqual, 1)
		Wish(t, stdout, ShouldEqual, "")
		Wish(t, stderr, ShouldEqual, Dedent(`
			reach: incorrect usage: 'not a command' is not a reach subcommand
		`))
	})
}
