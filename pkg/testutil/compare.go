package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
)

func EqualMessages(t *testing.T, expected, actual *model.Messages) {
	if expected == nil {
		assert.Equal(t, expected, actual)
		return
	}

	require.Equal(t, expected.Language, actual.Language)
	require.ElementsMatch(t, expected.Messages, actual.Messages)
}
