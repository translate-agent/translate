package convert

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

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
			name: "all values are provided",
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
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid "Hello, world!"
msgstr "Bonjour le monde!"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "multiline description",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting\nmultiline description",
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
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#. multiline description
#, fuzzy
msgid "Hello, world!"
msgstr "Bonjour le monde!"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "multiline msgid",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!\nvery long string\n",
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
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid ""
"Hello, world!\n"
"very long string\n"
msgstr "Bonjour le monde!"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "single msgid with newline",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!\n",
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
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid "Hello, world!\n"
msgstr "Bonjour le monde!"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "multiline msgstr",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!\nvery long string\n",
						Description: "A simple greeting", Fuzzy: true,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell", Fuzzy: true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid "Hello, world!"
msgstr ""
"Bonjour le monde!\n"
"very long string\n"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "qouted msgstr",
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
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid "Hello, world!"
msgstr "This is a \"quoted\" string"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "qouted msgid",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, \"world!\"",
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
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid "Hello, \"world!\""
msgstr "Bonjour le monde!"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "mixed fuzzy values",
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
						Fuzzy:       false,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#, fuzzy
msgid "Hello, world!"
msgstr "Bonjour le monde!"

#. A farewell
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "plural msgstr",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count} pomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] "Il y a %d pomme."
msgstr[1] "Il y a %d pommes."
`),
		},
		{
			name: "single message",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is apple.",
						Message:     "{Il y a pomme.}",
						Description: "apple counts",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. apple counts
#, fuzzy
msgid "There is apple."
msgstr "Il y a pomme."
`),
		},
		{
			name: "single message with new line",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is apple.",
						Message:     "{Il y \na pomme.}",
						Description: "apple counts",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. apple counts
#, fuzzy
msgid "There is apple."
msgstr ""
"Il y \n"
"a pomme.\n"
`),
		},
		{
			name: "single message with quoted message",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is apple.",
						Message:     "{Il y a \"pomme\".}",
						Description: "apple counts",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. apple counts
#, fuzzy
msgid "There is apple."
msgstr "Il y a \"pomme\"."
`),
		},
		{
			name: "plural msgstr with new line",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count}\npomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] ""
"Il y a %d\n"
"pomme.\n"
msgstr[1] "Il y a %d pommes."
`),
		},
		{
			name: "plural msgstr with new lines",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: "match {$count :number}\n" +
							"when 1 {Il y a {$count}\n" +
							"pomme.\n" +
							"one more line.}\n" +
							"when * {Il y a {$count}\n" +
							"pommes.}",
						Description: "apple counts",
						Fuzzy:       true,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] ""
"Il y a %d\n"
"pomme.\n"
"one more line.\n"
msgstr[1] ""
"Il y a %d\n"
"pommes.\n"
`),
		},
		{
			name: "missing fuzzy values",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Description: "A simple greeting"},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell"},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
msgid "Hello, world!"
msgstr "Bonjour le monde!"

#. A farewell
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "missing description",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!", Fuzzy: true},
					{ID: "Goodbye!", Message: "Au revoir!", Description: "A farewell", Fuzzy: true},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#, fuzzy
msgid "Hello, world!"
msgstr "Bonjour le monde!"

#. A farewell
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "missing description and fuzzy",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{ID: "Hello, world!", Message: "Bonjour le monde!"},
					{ID: "Goodbye!", Message: "Au revoir!"},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
msgid "Hello, world!"
msgstr "Bonjour le monde!"

msgid "Goodbye!"
msgstr "Au revoir!"
`),
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
			name: "valid input",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							#. a greeting
							msgid "Hello"
							msgstr "Hello, world!"

							#. a farewell
							#, fuzzy
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
			expected: model.Messages{
				Language: language.English,
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
			name: "multiline description",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							#. a greeting
							#. a greeting2
							msgid "Hello"
							msgstr ""
							"Hello, world!\n"
							"very long string\n"

							#. a farewell
							#, fuzzy
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "Hello, world!\nvery long string\n",
						Description: "a greeting\n a greeting2",
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
			name: "multiline msgid",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							#. a greeting
							#, fuzzy
							msgid ""
							"Hello\n"
							"Hello2\n"
							msgstr "Hello, world!"
			`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello\nHello2\n",
						Message:     "Hello, world!",
						Description: "a greeting",
						Fuzzy:       true,
					},
				},
			},
		},
		{
			name: "plural msgstr",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme."
							msgstr[1] "Il y a %d pommes."
			`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {Il y a {$count} pomme.}
when * {Il y a {$count} pommes.}
`,
						Description: "apple counts",
						Fuzzy:       false,
					},
				},
			},
		},
		{
			name: "plural msgstr with new line",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] ""
							"Il y a %d\n"
							"pomme.\n"
							msgstr[1] ""
							"Il y a %d\n"
							"pommes.\n"
			`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count}\npomme.}\nwhen * {Il y a {$count}\npommes.}\n",
						Description: "apple counts",
						Fuzzy:       false,
					},
				},
			},
		},
		{
			name: "multiline msgid_plural and msgid",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural ""
							"There are %d apples.\n"
							msgstr[0] ""
							"Il y a %d\n"
							"pomme.\n"
							msgstr[1] "Il y a %d pommes."
			`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.\n",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count}\npomme.}\nwhen * {Il y a {$count} pommes.}\n",
						Description: "apple counts",
						Fuzzy:       false,
					},
				},
			},
		},
		{
			name: "invalid input",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							#. a greeting
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
			if tt.expectedErr != nil {
				assert.Errorf(t, err, tt.expectedErr.Error())
				return
			}
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_TransformMessage(t *testing.T) {
	t.Parallel()

	n := gofakeit.IntRange(1, 3)

	lang := language.MustParse(gofakeit.LanguageBCP())

	msg := model.Messages{
		Language: lang,
		Messages: make([]model.Message, 0, n),
	}

	for i := 0; i < n; i++ {
		msg.Messages = append(msg.Messages, model.Message{
			ID:          gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
		},
		)
	}

	msgPot, err := ToPot(msg)
	require.NoError(t, err)

	restoredMsg, err := FromPot(msgPot)
	require.NoError(t, err)

	assert.Equal(t, msg, restoredMsg)
}
