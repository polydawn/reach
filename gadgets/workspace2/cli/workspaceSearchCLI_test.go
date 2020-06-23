package workspacecli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli"
	. "github.com/warpfork/go-wish"
)

func TestWorkspaceSearchCLI(t *testing.T) {
	t.Run("no workspace", func(t *testing.T) {
		WithTmpdir(t, func(dir string) {
			bangFile(dir+"/nothing_of_interest", "")

			exitCode, stdout, stderr := RunIntoBuffer(Cmd, "testapp", "--wd", dir, "search-workspace")
			Wish(t, exitCode, ShouldEqual, 0)
			Wish(t, stdout, ShouldEqual, Dedent(`
				workspace:        -none- (workspace not found (searched starting at "/tmp/reach-test/TestWorkspaceSearchCLI__no_workspace" and all the way up to the root without finding a '.timeless' dir))
			`))
			Wish(t, stderr, ShouldEqual, Dedent(""))
		})
	})
	t.Run("workspace but no module", func(t *testing.T) {
		WithTmpdir(t, func(dir string) {
			bangFile(dir+"/.timeless/", "")

			exitCode, stdout, stderr := RunIntoBuffer(Cmd, "testapp", "--wd", dir, "search-workspace")
			Wish(t, exitCode, ShouldEqual, 0)
			Wish(t, stdout, ShouldEqual, Dedent(`
				workspace:        /tmp/reach-test/TestWorkspaceSearchCLI__workspace_but_no_module
				parent workspace: -none- (workspace not found (searched starting at "/tmp/reach-test" and all the way up to the root without finding a '.timeless' dir))
				module root:      -none- (module not found (searched starting at "/tmp/reach-test/TestWorkspaceSearchCLI__workspace_but_no_module" and all the way up to the root without finding a 'module.tl' file))
			`))
			Wish(t, stderr, ShouldEqual, Dedent(""))
		})
	})
	t.Run("workspace with module at root", func(t *testing.T) {
		WithTmpdir(t, func(dir string) {
			bangFile(dir+"/.timeless/", "")
			bangFile(dir+"/module.tl", "")

			exitCode, stdout, stderr := RunIntoBuffer(Cmd, "testapp", "--wd", dir, "search-workspace")
			Wish(t, exitCode, ShouldEqual, 0)
			Wish(t, stdout, ShouldEqual, Dedent(`
				workspace:        /tmp/reach-test/TestWorkspaceSearchCLI__workspace_with_module_at_root
				parent workspace: -none- (workspace not found (searched starting at "/tmp/reach-test" and all the way up to the root without finding a '.timeless' dir))
				module root:      /tmp/reach-test/TestWorkspaceSearchCLI__workspace_with_module_at_root
				module name:      TODO: moduleName computation
			`))
			Wish(t, stderr, ShouldEqual, Dedent(""))
		})
	})
	t.Run("workspace with module deeper", func(t *testing.T) {
		WithTmpdir(t, func(dir string) {
			bangFile(dir+"/.timeless/", "")
			bangFile(dir+"/deeper/module.tl", "")

			exitCode, stdout, stderr := RunIntoBuffer(Cmd, "testapp", "--wd", dir+"/deeper", "search-workspace")
			Wish(t, exitCode, ShouldEqual, 0)
			Wish(t, stdout, ShouldEqual, Dedent(`
				workspace:        /tmp/reach-test/TestWorkspaceSearchCLI__workspace_with_module_deeper
				parent workspace: -none- (workspace not found (searched starting at "/tmp/reach-test" and all the way up to the root without finding a '.timeless' dir))
				module root:      /tmp/reach-test/TestWorkspaceSearchCLI__workspace_with_module_deeper/deeper
				module name:      TODO: moduleName computation
			`))
			Wish(t, stderr, ShouldEqual, Dedent(""))
		})
	})
	t.Run("workspace with module deeper with wd deeperer", func(t *testing.T) {
		WithTmpdir(t, func(dir string) {
			bangFile(dir+"/.timeless/", "")
			bangFile(dir+"/deeper/module.tl", "")
			bangFile(dir+"/deeper/deeperer/", "")

			exitCode, stdout, stderr := RunIntoBuffer(Cmd, "testapp", "--wd", dir+"/deeper/deeperer", "search-workspace")
			Wish(t, exitCode, ShouldEqual, 0)
			Wish(t, stdout, ShouldEqual, Dedent(`
				workspace:        /tmp/reach-test/TestWorkspaceSearchCLI__workspace_with_module_deeper_with_wd_deeperer
				parent workspace: -none- (workspace not found (searched starting at "/tmp/reach-test" and all the way up to the root without finding a '.timeless' dir))
				module root:      /tmp/reach-test/TestWorkspaceSearchCLI__workspace_with_module_deeper_with_wd_deeperer/deeper
				module name:      TODO: moduleName computation
			`))
			Wish(t, stderr, ShouldEqual, Dedent(""))
		})
	})
	t.Run("nested workspace with wd deeperer", func(t *testing.T) {
		WithTmpdir(t, func(dir string) {
			bangFile(dir+"/.timeless/", "")
			bangFile(dir+"/deeper/.timeless/", "")
			bangFile(dir+"/deeper/deeperer/", "")

			exitCode, stdout, stderr := RunIntoBuffer(Cmd, "testapp", "--wd", dir+"/deeper/deeperer", "search-workspace")
			Wish(t, exitCode, ShouldEqual, 0)
			Wish(t, stdout, ShouldEqual, Dedent(`
				workspace:        /tmp/reach-test/TestWorkspaceSearchCLI__nested_workspace_with_wd_deeperer/deeper
				parent workspace: /tmp/reach-test/TestWorkspaceSearchCLI__nested_workspace_with_wd_deeperer
				module root:      -none- (module not found (searched starting at "/tmp/reach-test/TestWorkspaceSearchCLI__nested_workspace_with_wd_deeperer/deeper/deeperer" and all the way up to the root without finding a 'module.tl' file))
			`))
			Wish(t, stderr, ShouldEqual, Dedent(""))
		})
	})
}

// about errors from the cli library:
//  we probably shouldn't test this here, but, I've noticed...
//  - this library emits args misuse info on the Writer, not the ErrWriter, which is insanely wrong.
//  - it **also** returns that error (in addition to already having flushed it to the Writer), so it's frustratingly likely you'll accidentally print it twice.

func RunIntoBuffer(cliCmd *cli.Command, args ...string) (int, string, string) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	cwd, err := os.Getwd() // not relevent to this test setup, but should be the norm for this app builder that we should be extracting.
	if err != nil {
		panic(err)
	}
	app := &cli.App{
		Name:      "testapp",
		Commands:  []*cli.Command{cliCmd},
		Writer:    stdout,
		ErrWriter: stderr,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:   "wd",
				Hidden: true,
				Value:  cwd,
			},
		},
	}
	if err := app.Run(args); err != nil {
		fmt.Fprintf(stderr, "testapp: error: %s\n", err)
		exitCode = 200
	}
	return exitCode, stdout.String(), stderr.String()
}

func WithTmpdir(t *testing.T, fn func(tmpDir string)) {
	tmpBase := "/tmp/reach-test/"
	err := os.MkdirAll(tmpBase, os.FileMode(0777)|os.ModeSticky)
	if err != nil {
		panic(err)
	}
	tmpdir := tmpBase + strings.ReplaceAll(t.Name(), "/", "__")
	err = os.MkdirAll(tmpdir, os.FileMode(0755))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)
	fn(tmpdir)
}

func bangFile(pth string, body string) {
	switch strings.HasSuffix(pth, "/") {
	case true:
		if err := os.MkdirAll(pth, os.FileMode(0755)); err != nil {
			panic(err)
		}
	case false:
		if err := os.MkdirAll(filepath.Dir(pth), os.FileMode(0755)); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(pth, []byte(body), os.FileMode(0644)); err != nil {
			panic(err)
		}
	}
}
