package testutil

import (
	"context"
	"io/ioutil"
	"os"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
)

func WithTmpdir(fn func(tmpDir fs.AbsolutePath)) {
	tmpBase := "/tmp/reach-test/"
	err := os.MkdirAll(tmpBase, os.FileMode(0777)|os.ModeSticky)
	if err != nil {
		panic(err)
	}

	tmpdir, err := ioutil.TempDir(tmpBase, "")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(tmpdir)
	fn(fs.MustAbsolutePath(tmpdir))
}

func WithClonedTmpdir(src fs.AbsolutePath, fn func(tmpDir fs.AbsolutePath)) {
	WithTmpdir(func(tmpDir fs.AbsolutePath) {
		if err := fsOp.Copy(context.Background(), osfs.New(src), osfs.New(tmpDir)); err != nil {
			panic(err)
		}
		fn(tmpDir)
	})
}

func WithCwdClonedTmpDir(src fs.AbsolutePath, fn func()) {
	WithClonedTmpdir(src, func(tmpDir fs.AbsolutePath) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		os.Chdir(tmpDir.String())
		defer os.Chdir(wd)
		fn()
	})
}

func GetCwdAbs() fs.AbsolutePath {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return fs.MustAbsolutePath(wd)
}
