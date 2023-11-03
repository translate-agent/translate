package convert

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

func Test_ToPot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected []byte
		input    model.Translation
	}{
		{
			name: "message with special chars",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     `{Bonjour \{user\} \| \\}`,
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
						Positions: []string{
							"examples/simple/example.clj:10",
							"examples/simple/example.clj:20",
						},
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#: examples/simple/example.clj:10
#: examples/simple/example.clj:20
#, fuzzy
msgid "Hello"
msgstr "Bonjour {user} | \"
`),
		},
		{
			name: "all values are provided",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
						Positions: []string{
							"examples/simple/example.clj:10",
							"examples/simple/example.clj:20",
						},
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
						Positions: []string{
							"examples/simple/example.clj:30",
							"examples/simple/example.clj:40",
						},
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
#. A simple greeting
#: examples/simple/example.clj:10
#: examples/simple/example.clj:20
#, fuzzy
msgid "Hello, world!"
msgstr "Bonjour le monde!"

#. A farewell
#: examples/simple/example.clj:30
#: examples/simple/example.clj:40
#, fuzzy
msgid "Goodbye!"
msgstr "Au revoir!"
`),
		},
		{
			name: "msgstr in curly braces",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "{Bonjour le monde!}",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
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
`),
		},
		{
			name: "msgstr with curly braces inside",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: `{Bonjour \{\} le monde!}`,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
msgid "Hello, world!"
msgstr "Bonjour {} le monde!"
`),
		},
		{
			name: "msgstr with slash inside",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: `{Bonjour \\ le monde!}`,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
msgid "Hello, world!"
msgstr "Bonjour \ le monde!"
`),
		},
		{
			name: "msgstr with pipe inside",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: `{Bonjour \| le monde!}`,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
msgid "Hello, world!"
msgstr "Bonjour | le monde!"
`),
		},
		{
			name: "msgstr with double pipe inside",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: `{Bonjour \|\| le monde!}`,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: en\n"
msgid "Hello, world!"
msgstr "Bonjour || le monde!"
`),
		},
		{
			name: "multiline description",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting\nmultiline description",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!\nvery long string\n",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!\n",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!\nvery long string\n",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
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
			name: "multiline msgstr in curly braces",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "{Bonjour le monde!\nvery long string}\n",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
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
`),
		},
		{
			name: "qouted msgstr",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "This is a \"quoted\" string",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
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
			name: "qouted msgstr in curly braces",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "{This is a \"quoted\" string}",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
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
`),
		},
		{
			name: "qouted msgid",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, \"world!\"",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "Bonjour le monde!",
						Description: "A simple greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusUntranslated,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count} pomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is apple.",
						Message:     "{Il y a pomme.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is apple.",
						Message:     "{Il y \na pomme.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is apple.",
						Message:     "{Il y a \"pomme\".}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count}\npomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
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
			input: model.Translation{
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
						Status:      model.MessageStatusFuzzy,
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
			name: "multiple plural messages",
			input: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count} pomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count} pomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: fr\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] "Il y a %d pomme."
msgstr[1] "Il y a %d pommes."

#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] "Il y a %d pomme."
msgstr[1] "Il y a %d pommes."
`),
		},
		{
			name: "multiple plural messages and original true",
			input: model.Translation{
				Language: language.French,
				Original: true,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count} pomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count} pomme.}\nwhen * {Il y a {$count} pommes.}",
						Description: "apple counts",
						Status:      model.MessageStatusFuzzy,
					},
				},
			},
			expected: []byte(`msgid ""
msgstr ""
"Language: fr\n"
#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] "Il y a %d pomme."
msgstr[1] "Il y a %d pommes."

#. apple counts
#, fuzzy
msgid "There is %d apple."
msgid_plural "There are %d apples."
msgstr[0] "Il y a %d pomme."
msgstr[1] "Il y a %d pommes."
`),
		},
		{
			name: "missing fuzzy values",
			input: model.Translation{
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
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: "Bonjour le monde!",
						Status:  model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Status:      model.MessageStatusFuzzy,
					},
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
			input: model.Translation{
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
			require.NoError(t, err)

			t.Logf("actual: %v", string(result))
			t.Logf("expect: %v", string(tt.expected))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromPot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr error
		input       []byte
		name        string
		expected    model.Translation
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
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Status:      model.MessageStatusUntranslated,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Status:      model.MessageStatusFuzzy,
					},
				},
			},
		},
		{
			name: "fuzzy param before empty id",
			input: []byte(`#, fuzzy
							msgid ""
							msgstr ""
							"Language: en\n"
							#. a greeting
							msgid "Hello"
							msgstr "Hello, world!"

							#. a farewell
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "msgid and msgstr empty headers",
			input: []byte(`#, fuzzy
							"Language: en\n"
							#. a greeting
							msgid "Hello"
							msgstr "Hello, world!"

							#. a farewell
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "if empty msgstr missing",
			input: []byte(`msgid ""
							"Language: en\n"
							#. a greeting
							msgid "Hello"
							msgstr "Hello, world!"

							#. a farewell
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Status:      model.MessageStatusUntranslated,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Status:      model.MessageStatusUntranslated,
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
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!\nvery long string\n}",
						Description: "a greeting\n a greeting2",
						Status:      model.MessageStatusUntranslated,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Status:      model.MessageStatusFuzzy,
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
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "Hello\nHello2\n",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Status:      model.MessageStatusFuzzy,
					},
				},
			},
		},
		{
			name: "Multiple headers",
			input: []byte(`msgid ""
						msgstr ""
						"Project-Id-Version: \n"
						"POT-Creation-Date: \n"
						"PO-Revision-Date: \n"
						"Last-Translator: \n"
						"Language-Team: \n"
						"Language: fr\n"
						"MIME-Version: 1.0\n"
						"Content-Type: text/plain; charset=UTF-8\n"
						"Content-Transfer-Encoding: 8bit\n"
						"X-Generator: Poedit 2.2\n"
						"Plural-Forms: nplurals=2; plural=(n > 1);\n"

						#: examples/simple/example.clj:10
						msgid "Greetings"
						msgstr "Bonjour"

						#: examples/simple/example.clj:20
						msgid "Please confirm your email"
						msgstr "Veuillez confirmer votre email"

						#: examples/simple/example.clj:30
						msgid "Welcome, %s!"
						msgstr "Bienvenue, %s!"

						#: examples/simple/example.clj:40
						#: examples/simple/example.clj:50
						msgid "product"
						msgid_plural "%s products"
						msgstr[0] "produit"
						msgstr[1] "%s produits"
			`),
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:        "Greetings",
						Message:   "{Bonjour}",
						Positions: []string{"examples/simple/example.clj:10"},
						Status:    model.MessageStatusUntranslated,
					},
					{
						ID:        "Please confirm your email",
						Message:   "{Veuillez confirmer votre email}",
						Positions: []string{"examples/simple/example.clj:20"},
						Status:    model.MessageStatusUntranslated,
					},
					{
						ID:        "Welcome, %s!",
						Message:   "{Bienvenue, %s!}",
						Positions: []string{"examples/simple/example.clj:30"},
						Status:    model.MessageStatusUntranslated,
					},
					{
						ID:       "product",
						PluralID: "%s products",
						Message: `match {$count :number}
when 1 {produit}
when * {%s produits}
`,
						Positions: []string{"examples/simple/example.clj:40", "examples/simple/example.clj:50"},
						Status:    model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr with simple msgstr",
			input: []byte(`msgid ""
							msgstr ""
							"Language: it\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme."
							msgstr[1] "Il y a %d pommes."

							msgid "hi"
							msgstr "ciao"
			`),
			expected: model.Translation{
				Language: language.Italian,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {Il y a {$count} pomme.}
when * {Il y a {$count} pommes.}
`,
						Description: "apple counts",
						Status:      model.MessageStatusUntranslated,
					},
					{
						ID:      "hi",
						Message: "{ciao}",
						Status:  model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr",
			input: []byte(`msgid ""
							msgstr ""
							"Language: fr\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme."
							msgstr[1] "Il y a %d pommes."
			`),
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {Il y a {$count} pomme.}
when * {Il y a {$count} pommes.}
`,
						Description: "apple counts",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr with new line",
			input: []byte(`msgid ""
							msgstr ""
							"Language: fr\n"
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
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count}\npomme.}\nwhen * {Il y a {$count}\npommes.}\n",
						Description: "apple counts",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "multiline msgid_plural and msgid",
			input: []byte(`msgid ""
							msgstr ""
							"Language: fr\n"
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
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "There is %d apple.",
						PluralID:    "There are %d apples.\n",
						Message:     "match {$count :number}\nwhen 1 {Il y a {$count}\npomme.}\nwhen * {Il y a {$count} pommes.}\n",
						Description: "apple counts",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "single msgstr with original lang",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							#. a greeting
							msgid "Hello"
							msgstr ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello}",
						Description: "a greeting",
						Status:      model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr with original lang",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] ""
							msgstr[1] ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {There is {$count} apple.}
when * {There are {$count} apples.}
`,
						Description: "apple counts",
						Status:      model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "msgstr plural without PluralForm header",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							msgid "message"
							msgid_plural "messages"
							msgstr[0] ""
							msgstr[1] ""

							#: superset/views/core.py:385
							#, python-format
							msgid "message2"
							msgstr ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "message",
						PluralID: "messages",
						Message: `match {$count :number}
when 1
when *
`,
						Description: "",
						Status:      model.MessageStatusUntranslated,
					},
					{
						ID:        "message2",
						Message:   "",
						Positions: []string{"superset/views/core.py:385"},
						Status:    model.MessageStatusUntranslated,
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
			expectedErr: errors.New("convert tokens to pot.Po: invalid po file: no messages found"),
		},
		{
			name: "msgid before empty msgstr is missing",
			input: []byte(`msgstr ""
							"Language: en\n"
							#. a greeting
							msgid "Hello"
							msgstr "Hello, world!"

							#. a farewell
							msgid "Goodbye"
							msgstr "Goodbye, world!"
			`),
			expectedErr: errors.New("convert tokens to pot.Po: get previous token: no previous token"),
		},
		{
			name: "msgid with curly braces inside",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							msgid "+ {%s} hello"
							msgstr ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "+ {%s} hello",
						Message: `{+ \{%s\} hello}`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "msgid with pipe inside",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							msgid "+ | hello"
							msgstr ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "+ | hello",
						Message: `{+ \| hello}`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "msgid with double pipe inside",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							msgid "+ || hello"
							msgstr ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "+ || hello",
						Message: `{+ \|\| hello}`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "msgid with slash inside",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							msgid "+ \ hello"
							msgstr ""
			`),
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "+ \\ hello",
						Message: `{+ \\ hello}`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr with curly braces",
			input: []byte(`msgid ""
							msgstr ""
							"Language: fr\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme {test}."
							msgstr[1] "Il y a %d pommes {tests}."
			`),
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {Il y a {$count} pomme \{test\}.}
when * {Il y a {$count} pommes \{tests\}.}
`,
						Status: model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr with pipe",
			input: []byte(`msgid ""
							msgstr ""
							"Language: fr\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme |."
							msgstr[1] "Il y a %d pommes |."
			`),
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {Il y a {$count} pomme \|.}
when * {Il y a {$count} pommes \|.}
`,
						Status: model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "plural msgstr with slash",
			input: []byte(`msgid ""
							msgstr ""
							"Language: fr\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme \."
							msgstr[1] "Il y a %d pommes \."
			`),
			expected: model.Translation{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "There is %d apple.",
						PluralID: "There are %d apples.",
						Message: `match {$count :number}
when 1 {Il y a {$count} pomme \\.}
when * {Il y a {$count} pommes \\.}
`,
						Status: model.MessageStatusUntranslated,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FromPot(tt.input, tt.expected.Original)
			if tt.expectedErr != nil {
				require.Errorf(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			testutil.EqualTranslations(t, &tt.expected, &result)
		})
	}
}

func Test_TransformMessage(t *testing.T) {
	t.Parallel()

	n := gofakeit.IntRange(1, 5)

	lang := language.MustParse(gofakeit.LanguageBCP())

	translation := model.Translation{
		Language: lang,
		Messages: make([]model.Message, 0, n),
		Original: false,
	}

	for i := 0; i < n; i++ {
		translation.Messages = append(translation.Messages, model.Message{
			ID:          gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
			Status:      model.MessageStatusUntranslated,
		},
		)
	}

	msgPot, err := ToPot(translation)
	require.NoError(t, err)

	restoredTranslation, err := FromPot(msgPot, translation.Original)
	require.NoError(t, err)

	assert.Equal(t, translation, restoredTranslation)
}
