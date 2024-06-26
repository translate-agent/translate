package convert

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

func Test_FromArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr error
		name        string
		input       []byte
		expected    model.Translation
	}{
		// Positive tests
		{
			name: "Message with special chars",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": "Message to greet the World"
				},
				"greeting": "Welcome {user} | \\ !",
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
			expected: model.Translation{
				Original: true,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "Hello World!",
						Description: `Message to greet the World`,
						Status:      model.MessageStatusTranslated,
					},
					{
						ID:      "greeting",
						Message: `Welcome \{user\} | \\ !`,
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "farewell",
						Message: `Goodbye friend`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "With locale",
			input: []byte(`
      {
        "@@locale": "lv",
        "title": "",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			expected: model.Translation{
				Language: language.Latvian,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     ``,
						Description: "Message to greet the World",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		// Negative tests
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
			expectedErr: errors.New("'description' expected type 'string', got unconvertible type 'map[string]interface {}', value: 'map[meaning:When you greet someone]'"), //nolint:lll
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
			expectedErr: errors.New("language: tag is not well-formed"),
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
			expectedErr: errors.New("unsupported value type 'map[string]interface {}' for key '@@locale'"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromArb(tt.input, &tt.expected.Original)
			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			testutil.EqualTranslations(t, &tt.expected, &actual)
		})
	}
}

func Test_ToArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected []byte
		input    model.Translation
	}{
		{
			name: "valid input",
			input: model.Translation{
				Language: language.French,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     `Hello World!`,
						Description: "Message to greet the World",
					},
					{
						ID:      "greeting",
						Message: `Welcome Sion`,
					},
				},
			},
			expected: []byte(`
	{
		"@@locale":"fr",
		"title":"Hello World!",
		"@title":{
			"description":"Message to greet the World"
		},
		"greeting":"Welcome Sion"
	}`),
		},
		{
			name: "Message with special chars",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "title",
						Message: `Hello World!`,
					},
					{
						ID:      "greeting",
						Message: `Welcome \{user\} | \\ !`,
					},
					{
						ID:      "farewell",
						Message: `Goodbye friend`,
					},
				},
			},
			expected: []byte(`
			{
				"@@locale":"en",
				"farewell":"Goodbye friend",
				"greeting":"Welcome {user} | \\ !",
				"title":"Hello World!"
			}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToArb(tt.input)
			require.NoError(t, err)

			assert.JSONEq(t, string(tt.expected), string(actual))
		})
	}
}
