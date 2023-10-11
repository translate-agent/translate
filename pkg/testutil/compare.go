package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
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
