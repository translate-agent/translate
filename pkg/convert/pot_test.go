package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_ToPot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		model    model.Messages
		expected []byte
	}{
		{
			name: "When all values are provided",
			model: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Description: "A simple greeting", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: true},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n#, fuzzy\nmsgid \"Hello, world!\"\nmsgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n#, fuzzy\nmsgid \"Goodbye!\"\nmsgstr \"Au revoir!\"\n"),
		},
		{
			name: "When fuzzy values are mixed",
			model: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Description: "A simple greeting", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: false},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n#, fuzzy\nmsgid \"Hello, world!\"\nmsgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\nmsgid \"Goodbye!\"\nmsgstr \"Au revoir!\"\n"),
		},
		{
			name: "When fuzzy values are missing",
			model: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Description: "A simple greeting"},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell"},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\nmsgid \"Hello, world!\"\nmsgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\nmsgid \"Goodbye!\"\nmsgstr \"Au revoir!\"\n"),
		},
		{
			name: "When description value is missing",
			model: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: true},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#, fuzzy\nmsgid \"Hello, world!\"\nmsgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n#, fuzzy\nmsgid \"Goodbye!\"\nmsgstr \"Au revoir!\"\n"),
		},
		{
			name: "When description and fuzzy values are missing",
			model: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!"},
					{ID: "Goodbye!", Message: "Au revoir!"},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"msgid \"Hello, world!\"\nmsgstr \"Bonjour le monde!\"\n" +
					"msgid \"Goodbye!\"\nmsgstr \"Au revoir!\"\n"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToPot(tt.model)

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromPot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expected    model.Messages
		expectedErr error
		input       []byte
	}{
		{
			name: "Valid input",
			input: []byte(`{
				"header": {
					"language": "en-US",
					"translator": "John Doe",
					"pluralForms": {"plural": "n != 1", "nplurals": 2}
				},
				"messages": [
					{
						"msgId": "Hello",
						"extractedComment": "a greeting",
						"flag": "",
						"msgStr": ["Hello, world!"]
					},
					{
						"msgId": "Goodbye",
						"extractedComment": "a farewell",
						"flag": "fuzzy",
						"msgStr": ["Goodbye, world!"]
					},
					{
						"msgIdPlural": "Apples",
						"extractedComment": "a plural message",
						"msgStr": ["Apple", "Apples"]
					}
				]
			}`),
			expected: model.Messages{
				Language: language.Make("en-US"),
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "Hello, world!",
						Description: "a greeting",
						Fuzzy:       false,
					},
					{
						ID:          "Goodbye",
						Message:     "Goodbye, world!",
						Description: "a farewell",
						Fuzzy:       true,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Invalid input",
			input: []byte(`{
				"header": {
					"language": "en-US",
					"translator": "John Doe",
					"pluralForms": {"plural": "n != 1", "nplurals": 2}
				},
				"messages": [
					{
						"msgId": "Hello",
						"extractedComment": "a greeting",
						"flag": "",
						"msgStr": ["Hello, world!"]
					},
					{
						"msgIdPlural": "Apples",
						"extractedComment": "a plural message",
						"msgStr": ["Apple", "Apples"]
					}
				]
			}`),
			expected: model.Messages{
				Language: language.Make("en-US"),
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "Hello, world!",
						Description: "a greeting",
						Fuzzy:       false,
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, actualErr := FromPot(tt.input)
			assert.Equal(t, tt.expectedErr, actualErr)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
