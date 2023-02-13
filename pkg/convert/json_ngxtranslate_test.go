package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
)

func Test_fromNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serialized []byte
		expected   model.Messages
		wantErr    bool
	}{
		{
			name:       "Not nested",
			serialized: []byte(`{"message":"example"}`),
			expected: model.Messages{
				Messages: []model.Message{
					{
						ID:      "message",
						Message: "example",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Nested normally",
			serialized: []byte(`{"message":{"example":"message1"}}`),
			expected: model.Messages{
				Messages: []model.Message{
					{
						ID:      "message.example",
						Message: "message1",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Nested with dot",
			serialized: []byte(`{"message.example":"message1"}`),
			expected: model.Messages{
				Messages: []model.Message{
					{
						ID:      "message.example",
						Message: "message1",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Nested mixed",
			serialized: []byte(`{"message.example":"message1","msg":{"example":"message2"}}`),
			expected: model.Messages{
				Messages: []model.Message{
					{
						ID:      "message.example",
						Message: "message1",
					},
					{
						ID:      "msg.example",
						Message: "message2",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FromNgxTranslate(tt.serialized)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_toNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		messages model.Messages
		expected []byte
		wantErr  bool
	}{
		{
			name: "Messages to NGX-Translate",
			messages: model.Messages{
				Messages: []model.Message{
					{
						ID:      "message",
						Message: "example",
					},
					{
						ID:      "message.example",
						Message: "message1",
					},
				},
			},
			expected: []byte(`{"message":"example","message.example":"message1"}`),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToNgxTranslate(tt.messages)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}
