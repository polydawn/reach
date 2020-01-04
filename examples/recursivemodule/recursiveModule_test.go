package hellomodule

import (
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/reach/examples/testutil"
)

func Test(t *testing.T) {
	if testing.Short() {
		t.Skipf("integration test -- not running with 'short' mode")
	}
	WithCwdClonedTmpDir(GetCwdAbs(), func() {
		exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge")
		Wish(t, exitCode, ShouldEqual, 0)
		Wish(t, stderr, ShouldEqual, Dedent(`
			module loaded
			module contains 5 steps
			module evaluation plan order:
			  - 01: step-first
			  - 02: submodule-jamboree
			  - 03: submodule-jamboree.bap
			  - 04: submodule-jamboree.boop
			  - 05: step-after
			imports pinned to hashes:
			  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
			  - "submodule-jamboree.image": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
			module eval complete.
			module exports:
		`))
		Wish(t, stdout, ShouldEqual, Dedent(`
			{
				"product": "tar:77k8uXWaArTqvyecQmjW9Xb3q3yMM9Mih842dA3HrrDW4Xs8uvU9kfipMWJKLQ5quZ"
			}
		`))
	})
}
