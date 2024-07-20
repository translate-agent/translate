package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	mf2 "go.expect.digital/mf2/parse"
	"go.expect.digital/translate/pkg/model"
)

func EqualTranslations(t *testing.T, want, got *model.Translation) {
	t.Helper()

	if want == nil {
		require.Equal(t, want, got)
	}

	require.Equal(t, want.Language, got.Language, "translation.language = %s, but want %s", got.Language, want.Language) //nolint:lll
	require.Equal(t, want.Original, got.Original, "translation.original = %t, but want %t", got.Original, want.Original) //nolint:lll
	require.ElementsMatch(t, want.Messages, got.Messages)
}

// EqualMF2Message compares two MessageFormat2 message ASTs.
func EqualMF2Message(t *testing.T, want, got string) {
	t.Helper()

	wantAST, err := mf2.Parse(want)
	require.NoError(t, err)

	gotAST, err := mf2.Parse(got)
	require.NoError(t, err)

	require.Equal(t, wantAST, gotAST)
}
