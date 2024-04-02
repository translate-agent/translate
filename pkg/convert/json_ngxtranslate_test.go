package convert

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
)

func Test_FromNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       []byte
		expectedErr error
		expected    model.Translation
	}{
		// Positive tests
		{
			name:  "Not nested",
			input: []byte(`{"message":"example"}`),
			expected: model.Translation{
				Original: true,
				Messages: []model.Message{
					{
						ID:      "message",
						Message: `example`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name:  "Message with special chars",
			input: []byte(`{"message":"Order #{Id} has been canceled for {ClientName} | \\"}`),
			expected: model.Translation{
				Original: true,
				Messages: []model.Message{
					{
						ID:      "message",
						Message: `Order #\{Id\} has been canceled for \{ClientName\} | \\`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name:  "Nested normally",
			input: []byte(`{"message":{"example":"message1"}}`),
			expected: model.Translation{
				Original: false,
				Messages: []model.Message{
					{
						ID:      "message.example",
						Message: `message1`,
						Status:  model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name:  "Nested with dot",
			input: []byte(`{"message.example":""}`),
			expected: model.Translation{
				Original: false,
				Messages: []model.Message{
					{
						ID:      "message.example",
						Message: ``,
						Status:  model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name:  "Nested mixed",
			input: []byte(`{"message.example":"message1","msg":{"example":"message2"}}`),
			expected: model.Translation{
				Original: true,
				Messages: []model.Message{
					{
						ID:      "message.example",
						Message: `message1`,
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "msg.example",
						Message: `message2`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		// Negative tests
		{
			name:        "Unsupported value type",
			input:       []byte(`{"message": 1.0}`),
			expectedErr: errors.New("unsupported value type"),
		},
		{
			name:        "Invalid JSON",
			input:       []byte(`{"message": "example"`),
			expectedErr: errors.New("unmarshal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromNgxTranslate(tt.input, &tt.expected.Original)
			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			testutil.EqualTranslations(t, &tt.expected, &actual)
		})
	}
}

func Test_ToNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expected []byte
		name     string
		input    model.Translation
	}{
		{
			name: "valid input",
			input: model.Translation{
				Messages: []model.Message{
					{
						ID:      "message",
						Message: `example`,
					},
					{
						ID:      "message.example",
						Message: `message1`,
					},
				},
			},
			expected: []byte(`{"message":"example","message.example":"message1"}`),
		},
		{
			name: "message with special chars",
			input: model.Translation{
				Messages: []model.Message{
					{
						ID:      "message",
						Message: `Welcome \{user\} | \\ !`,
					},
				},
			},
			expected: []byte(`{"message":"Welcome {user} | \\ !"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToNgxTranslate(tt.input)

			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}
