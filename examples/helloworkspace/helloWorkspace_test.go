package helloworkspace

import (
	"os"
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/stellar/examples/testutil"
)

/*
	- proj-foo depends on only one thing, already released in the catalog.
	- proj-bar depends on the candidate release of proj-foo!
	- proj-baz depends on an already released version of proj-foo.
*/

func TestEmergeOutsideModule(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("stellar", "emerge")
	Wish(t, exitCode, ShouldEqual, 1)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		stellar: no module found -- run this command in a module dir (e.g contains module.tl file), or specify a path to one
	`))
}

func TestEmergeInModule(t *testing.T) {
	WithCwdClonedTmpDir(GetCwdAbs(), func() {
		os.Chdir("example.org/proj-foo")
		exitCode, stdout, stderr := RunIntoBuffer("stellar", "emerge")
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
				exitCode, stdout, stderr := RunIntoBuffer("stellar", "emerge")
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
				exitCode, stdout, stderr := RunIntoBuffer("stellar", "emerge")
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

// TODO TestEmergeRecursion
// two modes:
//  - default: halt with warning, because no such release found while resolving pins.
//  - if recursion allowed: automatically fall into commission mode, and thus build the proj-bar candidate.
// the halt mode will currently happen; recursion mode needs a bunch o' code :)

func TestLint(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("stellar", "catalog", "lint", ".timeless/catalogs/upstream")
	Wish(t, exitCode, ShouldEqual, 0)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		0 total warnings
	`))
}
