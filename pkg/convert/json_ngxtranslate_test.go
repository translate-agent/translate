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
		name    string
		input   []byte
		wantErr error
		want    model.Translation
	}{
		// Positive tests
		{
			name:  "Not nested",
			input: []byte(`{"message":"example"}`),
			want: model.Translation{
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
			want: model.Translation{
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
			want: model.Translation{
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
			want: model.Translation{
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
			want: model.Translation{
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
			name:    "Unsupported value type",
			input:   []byte(`{"message": 1.0}`),
			wantErr: errors.New("unsupported value type"),
		},
		{
			name:    "Invalid JSON",
			input:   []byte(`{"message": "example"`),
			wantErr: errors.New("unmarshal"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromNgxTranslate(test.input, &test.want.Original)
			if test.wantErr != nil {
				require.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			require.NoError(t, err)
			testutil.EqualTranslations(t, &test.want, &got)
		})
	}
}

func Test_ToNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want  []byte
		name  string
		input model.Translation
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
			want: []byte(`{"message":"example","message.example":"message1"}`),
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
			want: []byte(`{"message":"Welcome {user} | \\ !"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToNgxTranslate(test.input)

			require.NoError(t, err)

			assert.Equal(t, test.want, got)
		})
	}
}
