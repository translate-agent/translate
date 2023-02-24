package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_ToPot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    model.Messages
		expected []byte
	}{
		{
			name: "When all values are provided",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Fuzzy:       true,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n#, fuzzy\nmsgid \"Hello, world!\"\nmsgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n#, fuzzy\nmsgid \"Goodbye!\"\nmsgstr \"Au revoir!\"\n"),
		},
		{
			name: "When msgid is multiline",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!\\n\"\n\"very long string\\n",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Fuzzy:       true,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n#, " +
					"fuzzy\n" +
					"msgid \"\" \n\"Hello, world!\\n\"\n\"very long string\\n\"\n" +
					"msgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n" +
					"#, fuzzy\nmsgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When msgstr is multiline",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!\\n\"\n\"very long string\\n",
						Description: "A simple greeting", Fuzzy: true,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell", Fuzzy: true,
					},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n" +
					"#, fuzzy\n" +
					"msgid \"Hello, world!\"\n" +
					"msgstr \"\" \n\"Bonjour le monde!\\n\"\n\"very long string\\n\"\n" +
					"#. A farewell\n" +
					"#, fuzzy\nmsgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When msgstr value is qouted",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "This is a \"quoted\" string",
						Description: "A simple greeting",
						Fuzzy:       true,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n" +
					"#, fuzzy\nmsgid \"Hello, world!\"\n" +
					"msgstr \"This is a \\\"quoted\\\" string\"\n" +
					"#. A farewell\n" +
					"#, fuzzy\nmsgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When msgid value is qouted",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, \"world!\"", Message: "Bonjour le monde!", Description: "A simple greeting", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: true},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n" +
					"#, fuzzy\n" +
					"msgid \"Hello, \\\"world!\\\"\"\n" +
					"msgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n" +
					"#, fuzzy\nmsgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When fuzzy values are mixed",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Description: "A simple greeting", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: false},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n" +
					"#, fuzzy\nmsgid \"Hello, world!\"\n" +
					"msgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n" +
					"msgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When fuzzy values are missing",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Description: "A simple greeting"},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell"},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#. A simple greeting\n" +
					"msgid \"Hello, world!\"\n" +
					"msgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n" +
					"msgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When description value is missing",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: true},
				},
			},
			expected: []byte(
				"\"Language: en\n" +
					"#, fuzzy\n" +
					"msgid \"Hello, world!\"\n" +
					"msgstr \"Bonjour le monde!\"\n" +
					"#. A farewell\n" +
					"#, fuzzy\nmsgid \"Goodbye!\"\n" +
					"msgstr \"Au revoir!\"\n"),
		},
		{
			name: "When description and fuzzy values are missing",
			input: model.Messages{
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

			result, err := ToPot(tt.input)

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
			input: []byte(`# Language: en-US
							#. "a greeting"
							#, ""
							msgid "Hello"
							msgstr "Hello, world!"
							
							#. "a farewell"
							#, "fuzzy"
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
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
		},
		{
			name: "Invalid input",
			input: []byte(`# Language: en-US
							#. "a greeting"
							#, ""
							msgid 323344
			`),
			expectedErr: fmt.Errorf("convert tokens to pot.Po: invalid po file: no messages found"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := FromPot(tt.input)
			if err != nil {
				assert.Equal(t, tt.expectedErr, fmt.Errorf(err.Error()))
				return
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}
