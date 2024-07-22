package convert

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/expect"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromNgxTranslate(tt.input, &tt.want.Original)
			if tt.wantErr != nil {
				expect.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			expect.NoError(t, err)
			testutil.EqualTranslations(t, &tt.want, &got)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToNgxTranslate(tt.input)

			expect.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
