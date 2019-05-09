package testutil

import (
	"bytes"
	"context"

	"go.polydawn.net/reach/cmd/reach/app"
)

func RunIntoBuffer(args ...string) (int, string, string) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := reach.Main(context.Background(), args, nil, stdout, stderr)
	return exitCode, stdout.String(), stderr.String()
}
