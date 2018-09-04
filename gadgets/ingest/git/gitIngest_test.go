package gitingest

import (
	"context"
	"testing"

	"go.polydawn.net/go-timeless-api"
)

// TestPrintfingly emits some logs -- but is not a real test and cannot fail.
// Refraining from making a real test here because git is a bear;
// we don't accept that the source tree is necessarly a git repo itself;
// and we don't want to bother with other fixture setup (yet).
func TestPrintfingly(t *testing.T) {
	wareID, wareSourcing, err := Resolve(context.Background(), api.ImportRef_Ingest{"git", "../..:HEAD"})
	t.Logf("%v\n%v\n%v\n\n", wareID, wareSourcing, err)
}
