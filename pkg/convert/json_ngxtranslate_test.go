package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
)

func Test_FromNgxTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serialized  []byte
		expectedErr error
		expected    model.Messages
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
		},
		{
			name:        "Unsupported value type",
			serialized:  []byte(`{"message": 1.0}`),
			expectedErr: fmt.Errorf("traverse ngx-translate: unsupported value type %T for key %s", 1.0, "message"),
			expected:    model.Messages{},
		},
		{
			name:        "Invalid JSON",
			serialized:  []byte(`{"message": "example"`),
			expectedErr: fmt.Errorf("unmarshal from ngx-translate to model.Messages: unexpected end of JSON input"),
			expected:    model.Messages{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FromNgxTranslate(tt.serialized)
			if err != nil {
				assert.Equal(t, tt.expectedErr, fmt.Errorf(err.Error()))
				return
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_ToNgxTranslate(t *testing.T) {
	t.Parallel()

	messages := model.Messages{
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
	}

	expected := []byte(`{"message":"example","message.example":"message1"}`)
	result, err := ToNgxTranslate(messages)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expected, result)
}
