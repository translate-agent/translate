package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
)

func EqualMessages(t *testing.T, expected, actual *model.Messages) {
	if expected == nil {
		require.Equal(t, expected, actual)
	}

	require.Equal(t, expected.Language, actual.Language)
	require.ElementsMatch(t, expected.Messages, actual.Messages)
}
