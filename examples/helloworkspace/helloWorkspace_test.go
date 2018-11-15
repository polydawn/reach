package helloworkspace

import (
	"os"
	"testing"

	. "github.com/warpfork/go-wish"

	. "go.polydawn.net/stellar/examples/testutil"
)

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
		os.Chdir("proj-bar")
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
	})
}

func TestLint(t *testing.T) {
	exitCode, stdout, stderr := RunIntoBuffer("stellar", "catalog", "lint", ".timeless/catalogs/upstream")
	Wish(t, exitCode, ShouldEqual, 0)
	Wish(t, stdout, ShouldEqual, "")
	Wish(t, stderr, ShouldEqual, Dedent(`
		0 total warnings
	`))
}
