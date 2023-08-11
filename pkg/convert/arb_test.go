package convert

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr error
		name        string
		input       []byte
		expected    model.Messages
	}{
		{
			name: "Combination of messages",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": "Message to greet the World"
				},
				"greeting": "Welcome {user}!",
				"@greeting": {
					"placeholders": {
						"user": {
							"type": "string",
							"example": "Bob"
						}
					}
				},
				"farewell": "Goodbye friend"
			}`),
			expected: model.Messages{
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "{Hello World!}",
						Description: "Message to greet the World",
					},
					{
						ID:      "greeting",
						Message: "{Welcome {user}!}",
					},
					{
						ID:      "farewell",
						Message: "{Goodbye friend}",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Message in curly braces",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": "Message to greet the World"
				},
				"greeting": "Welcome {user}!",
				"@greeting": {
					"placeholders": {
						"user": {
							"type": "string",
							"example": "Bob"
						}
					}
				},
				"farewell": "Goodbye friend"
			}`),
			expected: model.Messages{
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "{Hello World!}",
						Description: "Message to greet the World",
					},
					{
						ID:      "greeting",
						Message: "{Welcome {user}!}",
					},
					{
						ID:      "farewell",
						Message: "{Goodbye friend}",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Wrong value type for @title",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": "Message to greet the World"
			}`),
			expectedErr: errors.New("expected a map, got 'string'"),
		},
		{
			name: "Wrong value type for greeting key",
			input: []byte(`
			{
				"title": "Hello World!",
				"greeting": {
					"description": "Needed for greeting"
				}
			}`),
			expectedErr: errors.New("unsupported value type 'map[string]interface {}' for key 'greeting'"),
		},
		{
			name: "Wrong value type for description key",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": {
						"meaning": "When you greet someone"
					}
				}
			}`),
			expectedErr: errors.New("'Description' expected type 'string', got unconvertible type 'map[string]interface {}'"),
		},
		{
			name: "With locale",
			input: []byte(`
      {
        "@@locale": "en",
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "{Hello World!}",
						Description: "Message to greet the World",
					},
				},
				Original: false,
			},
			expectedErr: nil,
		},
		{
			name: "With malformed locale",
			input: []byte(`
      {
        "@@locale": "asd-gh-jk",
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			expectedErr: fmt.Errorf("language: tag is not well-formed"),
		},
		{
			name: "With wrong value type for locale",
			input: []byte(`
      {
        "@@locale": {
          "tag": "fr-FR"
        },
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			expectedErr: fmt.Errorf("unsupported value type 'map[string]interface {}' for key '@@locale'"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromArb(tt.input, false)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected.Language, actual.Language)
			assert.ElementsMatch(t, tt.expected.Messages, actual.Messages)
		})
	}
}

func Test_ToArb(t *testing.T) {
	t.Parallel()

	messages := model.Messages{
		Language: language.French,
		Messages: []model.Message{
			{
				ID:          "title",
				Message:     "{Hello World!}",
				Description: "Message to greet the World",
			},
			{
				ID:      "greeting",
				Message: "{Welcome Sion}",
			},
		},
	}

	expected := []byte(`
	{
		"@@locale":"fr",
		"title":"Hello World!",
		"@title":{
			"description":"Message to greet the World"
		},
		"greeting":"Welcome Sion"
	}`)

	actual, err := ToArb(messages)
	if !assert.NoError(t, err) {
		return
	}

	assert.JSONEq(t, string(expected), string(actual))
}
