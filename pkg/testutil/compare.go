package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	mf2 "go.expect.digital/mf2/parse"
)

// EqualMF2Message compares two MessageFormat2 message ASTs.
func EqualMF2Message(t *testing.T, want, got string) {
	t.Helper()

	wantAST, err := mf2.Parse(want)
	if err != nil {
		t.Error(err)
		return
	}

	gotAST, err := mf2.Parse(got)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, wantAST, gotAST)
}
