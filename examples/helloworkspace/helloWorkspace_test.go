package helloworkspace

import (
	"os"
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/reach/examples/testutil"
)

/*
	- proj-foo depends on only one thing, already released in the catalog.
	- proj-bar depends on the candidate release of proj-foo!
	- proj-baz depends on an already released version of proj-foo.
*/

func TestEmergeOutsideModule(t *testing.T) {
	if testing.Short() {
		t.Skipf("integration test -- not running with 'short' mode")
	}
	WithCwdClonedTmpDir(GetCwdAbs(), func() {
		exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge")
		Wish(t, exitCode, ShouldEqual, 1)
		Wish(t, stdout, ShouldEqual, "")
		Wish(t, stderr, ShouldEqual, Dedent(`
			reach: no module found
		`))
	})
}

func TestEmergeInModule(t *testing.T) {
	if testing.Short() {
		t.Skipf("integration test -- not running with 'short' mode")
	}
	WithCwdClonedTmpDir(GetCwdAbs(), func() {
		os.Chdir("example.org/proj-foo")
		exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge")
		Wish(t, exitCode, ShouldEqual, 0)
		Wish(t, stderr, ShouldEqual, Dedent(`
			module loaded
			module contains 1 steps
			module evaluation plan order:
			  - 01: main
			imports pinned to hashes:
			  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
			module eval complete.
			module exports:
		`))
		Wish(t, stdout, ShouldEqual, Dedent(`
			{
				"wowslot": "tar:89LoLzgAYkndYpNQC7H94eR6tU6F4EWy2yFGouDCQz1cx9JpYmEPyDm2YWwYTGDvPv"
			}
		`))

		t.Run("a candidate release should be recorded", func(t *testing.T) {
			// n.b. we could test the presence of the files here, but
			//  we'll choose not to.  That's not really a publicly exposed detail,
			//   and our next functional tests confirm everything quite clearly.

			t.Run("another module can consume the candidate", func(t *testing.T) {
				os.Chdir("../proj-bar")
				exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge")
				Wish(t, exitCode, ShouldEqual, 0)
				Wish(t, stderr, ShouldEqual, Dedent(`
					module loaded
					module contains 1 steps
					module evaluation plan order:
					  - 01: main
					imports pinned to hashes:
					  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
					  - "pipe": tar:89LoLzgAYkndYpNQC7H94eR6tU6F4EWy2yFGouDCQz1cx9JpYmEPyDm2YWwYTGDvPv
					module eval complete.
					module exports:
				`))
				Wish(t, stdout, ShouldEqual, Dedent(`
					{}
				`))
			})

			t.Run("the non-candidate release is still visible", func(t *testing.T) {
				os.Chdir("../proj-baz")
				exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge")
				Wish(t, exitCode, ShouldEqual, 0)
				Wish(t, stderr, ShouldEqual, Dedent(`
					module loaded
					module contains 1 steps
					module evaluation plan order:
					  - 01: main
					imports pinned to hashes:
					  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
					  - "pipe": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
					module eval complete.
					module exports:
				`))
				Wish(t, stdout, ShouldEqual, Dedent(`
					{}
				`))
			})
		})
	})
}

func TestEmergeViaModuleArg(t *testing.T) {
	if testing.Short() {
		t.Skipf("integration test -- not running with 'short' mode")
	}
	WithCwdClonedTmpDir(GetCwdAbs(), func() {
		exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge", "example.org/proj-foo")
		Wish(t, exitCode, ShouldEqual, 0)
		Wish(t, stderr, ShouldEqual, Dedent(`
			module loaded
			module contains 1 steps
			module evaluation plan order:
			  - 01: main
			imports pinned to hashes:
			  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
			module eval complete.
			module exports:
		`))
		Wish(t, stdout, ShouldEqual, Dedent(`
			{
				"wowslot": "tar:89LoLzgAYkndYpNQC7H94eR6tU6F4EWy2yFGouDCQz1cx9JpYmEPyDm2YWwYTGDvPv"
			}
		`))

		t.Run("a candidate release should be recorded", func(t *testing.T) {
			// n.b. we could test the presence of the files here, but
			//  we'll choose not to.  That's not really a publicly exposed detail,
			//   and our next functional tests confirm everything quite clearly.

			t.Run("another module can consume the candidate", func(t *testing.T) {
				exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge", "example.org/proj-bar")
				Wish(t, exitCode, ShouldEqual, 0)
				Wish(t, stderr, ShouldEqual, Dedent(`
					module loaded
					module contains 1 steps
					module evaluation plan order:
					  - 01: main
					imports pinned to hashes:
					  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
					  - "pipe": tar:89LoLzgAYkndYpNQC7H94eR6tU6F4EWy2yFGouDCQz1cx9JpYmEPyDm2YWwYTGDvPv
					module eval complete.
					module exports:
				`))
				Wish(t, stdout, ShouldEqual, Dedent(`
					{}
				`))
			})

			t.Run("the non-candidate release is still visible", func(t *testing.T) {
				exitCode, stdout, stderr := RunIntoBuffer("reach", "emerge", "example.org/proj-baz")
				Wish(t, exitCode, ShouldEqual, 0)
				Wish(t, stderr, ShouldEqual, Dedent(`
					module loaded
					module contains 1 steps
					module evaluation plan order:
					  - 01: main
					imports pinned to hashes:
					  - "base": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
					  - "pipe": tar:6q7G4hWr283FpTa5Lf8heVqw9t97b5VoMU6AGszuBYAz9EzQdeHVFAou7c4W9vFcQ6
					module eval complete.
					module exports:
				`))
				Wish(t, stdout, ShouldEqual, Dedent(`
					{}
				`))
			})
		})
	})
}

func TestEmergeRecursion(t *testing.T) {
	if testing.Short() {
		t.Skipf("integration test -- not running with 'short' mode")
	}
	t.Run("one recursion (when not enabled) should fail", func(t *testing.T) {
		WithCwdClonedTmpDir(GetCwdAbs(), func() {
			exitCode, _, _ := RunIntoBuffer("reach", "emerge", "example.org/proj-bar")
			Wish(t, exitCode, ShouldEqual, 1)
		})
	})
	t.Run("one recursion should succeed", func(t *testing.T) {
		WithCwdClonedTmpDir(GetCwdAbs(), func() {
			exitCode, _, _ := RunIntoBuffer("reach", "emerge", "-r", "example.org/proj-bar")
			Wish(t, exitCode, ShouldEqual, 0)
		})
	})
	// TODO: more recursion tests
	//  (but possibly start another example dir?  this one is complex enough.)
}

func TestLint(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("reach", "catalog", "lint", ".timeless/catalog")
	Wish(t, exitCode, ShouldEqual, 0)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		0 total warnings
	`))
}
