package stellar_test

import (
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/stellar/examples/testutil"
)

func TestNoargsHelptext(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("stellar")
	Wish(t, exitCode, ShouldEqual, 0)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		NAME:
		   stellar - sidereal repeatr

		USAGE:
		   Stellar builds modules of repeatr operations, stages releases of the results, and can commission builds of entire generations of atomic releases from many modules.

		COMMANDS:
		     emerge   evaluate a pipeline, logging intermediate results and reporting final exports
		     catalog  catalog subcommands help maintain the release catalog info tree
		     help, h  Shows a list of commands or help for one command

		GLOBAL OPTIONS:
		   --help, -h     show help
		   --version, -v  print the version
	`))
}

func TestWrongCommandHelptext(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("stellar", "not a command")
	Wish(t, exitCode, ShouldEqual, 1)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		stellar: incorrect usage: 'not a command' is not a stellar subcommand
	`))

	t.Run("also when asking for help", func(t *testing.T) {
		exitCode, stdout, stderr := RunIntoBuffer("stellar", "not a command", "-h")
		Wish(t, exitCode, ShouldEqual, 1)
		Wish(t, stdout, ShouldEqual, "")
		Wish(t, stderr, ShouldEqual, Dedent(`
			stellar: incorrect usage: 'not a command' is not a stellar subcommand
		`))
	})
}
