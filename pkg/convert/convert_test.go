package convert

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
)

func Test_getStatus(t *testing.T) {
	t.Parallel()

	type input struct {
		msg             string
		original, fuzzy bool
	}

	tests := []struct {
		name  string
		input struct {
			msg             string
			original, fuzzy bool
		}
		expected model.MessageStatus
	}{
		{
			name:     "Original is true",
			expected: model.MessageStatusTranslated,
			input: input{
				msg:      gofakeit.Name(),
				original: true,
				fuzzy:    false,
			},
		},
		{
			name:     "Original is false not empty message",
			expected: model.MessageStatusTranslated,
			input: input{
				msg:      gofakeit.Name(),
				original: false,
				fuzzy:    false,
			},
		},
		{
			name:     "Original is false empty message",
			expected: model.MessageStatusUntranslated,
			input: input{
				msg:      "",
				original: false,
				fuzzy:    false,
			},
		},
		{
			name:     "Fuzzy is true",
			expected: model.MessageStatusFuzzy,
			input: input{
				msg:      gofakeit.Name(),
				original: false,
				fuzzy:    true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := getStatus(tt.input.msg, tt.input.original, tt.input.fuzzy)
			require.Equal(t, tt.expected, status)
		})
	}
}
