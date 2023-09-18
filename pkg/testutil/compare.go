package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
)

func EqualMessages(t *testing.T, expected, actual *model.Messages) {
	t.Helper()

	if expected == nil {
		require.Equal(t, expected, actual)
	}

	require.Equal(t, expected.Language, actual.Language, "messages.language = %s, but want %s", actual.Language, expected.Language) //nolint:lll
	require.Equal(t, expected.Original, actual.Original, "messages.original = %t, but want %t", actual.Original, expected.Original) //nolint:lll
	require.ElementsMatch(t, expected.Messages, actual.Messages)
}
