package testutil

import (
	"bytes"
	"context"

	"go.polydawn.net/stellar/cmd/stellar/app"
)

func RunIntoBuffer(args ...string) (int, string, string) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := stellar.Main(context.Background(), args, nil, stdout, stderr)
	return exitCode, stdout.String(), stderr.String()
}
