package convert

import (
	"bytes"
	"reflect"
	"slices"
	"testing"

	"go.expect.digital/translate/pkg/model"
)

func Test_FromNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr string
		input   []byte
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
			wantErr: "traverse ngx-translate: unsupported value type float64 for key message",
		},
		{
			name:    "Invalid JSON",
			input:   []byte(`{"message": "example"`),
			wantErr: "unmarshal from ngx-translate to model.Translation: unexpected end of JSON input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromNgxTranslate(test.input, &test.want.Original)
			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			cmp := func(a, b model.Message) int {
				switch {
				default:
					return 0
				case a.ID < b.ID:
					return -1
				case b.ID < a.ID:
					return 1
				}
			}

			slices.SortFunc(test.want.Messages, cmp)
			slices.SortFunc(got.Messages, cmp)

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("\nwant %v\ngot  %v", test.want, got)
			}
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
			if err != nil {
				t.Error(err)
				return
			}

			if !bytes.Equal(test.want, got) {
				t.Errorf("want ngxtranslate '%s', got '%s'", string(test.want), string(got))
			}
		})
	}
}
