package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	mf2 "go.expect.digital/mf2/parse"
	"go.expect.digital/translate/pkg/testutil/expect"
)

// EqualMF2Message compares two MessageFormat2 message ASTs.
func EqualMF2Message(t *testing.T, want, got string) {
	t.Helper()

	wantAST, err := mf2.Parse(want)
	expect.NoError(t, err)

	gotAST, err := mf2.Parse(got)
	expect.NoError(t, err)

	require.Equal(t, wantAST, gotAST)
}
