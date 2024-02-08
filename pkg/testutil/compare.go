package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	mf2 "go.expect.digital/mf2/parse"
	"go.expect.digital/translate/pkg/model"
)

func EqualTranslations(t *testing.T, expected, actual *model.Translation) {
	t.Helper()

	if expected == nil {
		require.Equal(t, expected, actual)
	}

	require.Equal(t, expected.Language, actual.Language, "translation.language = %s, but want %s", actual.Language, expected.Language) //nolint:lll
	require.Equal(t, expected.Original, actual.Original, "translation.original = %t, but want %t", actual.Original, expected.Original) //nolint:lll
	require.ElementsMatch(t, expected.Messages, actual.Messages)
}

// EqualMF2Message compares two MessageFormat2 message ASTs.
func EqualMF2Message(t *testing.T, expected, actual string) {
	t.Helper()

	expectedAST, err := mf2.Parse(expected)
	require.NoError(t, err)

	actualAST, err := mf2.Parse(actual)
	require.NoError(t, err)

	require.Equal(t, expectedAST, actualAST)
}
