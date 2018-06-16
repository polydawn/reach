package hellomodule

import (
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/stellar/examples/testutil"
)

func Test(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("stellar", "emerge")
	Wish(t, exitCode, ShouldEqual, 0)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		workspace loaded
		module contains 1 steps
		module evaluation plan order:
		  - 01: main
	`))
}
