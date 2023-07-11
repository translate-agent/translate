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
		expected []byte
		input    model.Messages
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
						Positions: []string{
							"examples/simple/example.clj:10",
							"examples/simple/example.clj:20",
						},
						Fuzzy: true,
					},
					{
						ID:          "Goodbye!",
						Message:     "Au revoir!",
						Description: "A farewell",
						Positions: []string{
							"examples/simple/example.clj:30",
							"examples/simple/example.clj:40",
						},
						Fuzzy: true,
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
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "{Bonjour le monde!}",
						Description: "A simple greeting",
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
			name: "multiline msgstr in curly braces",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "{Bonjour le monde!\nvery long string}\n",
						Description: "A simple greeting", Fuzzy: true,
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
			name: "qouted msgstr in curly braces",
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello, world!",
						Message:     "{This is a \"quoted\" string}",
						Description: "A simple greeting",
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
		expectedErr error
		input       []byte
		name        string
		expected    model.Messages
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
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Fuzzy:       false,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Fuzzy:       true,
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
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Fuzzy:       true,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Fuzzy:       false,
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
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Fuzzy:       true,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Fuzzy:       false,
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
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Hello",
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Fuzzy:       false,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
						Description: "a farewell",
						Fuzzy:       false,
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
						Message:     "{Hello, world!\nvery long string\n}",
						Description: "a greeting\n a greeting2",
						Fuzzy:       false,
					},
					{
						ID:          "Goodbye",
						Message:     "{Goodbye, world!}",
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
						Message:     "{Hello, world!}",
						Description: "a greeting",
						Fuzzy:       true,
					},
				},
			},
		},
		{
			name: "all possible headers with plural messag",
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
						msgstr[1] "%s produits"`),
			expected: model.Messages{
				Language: language.French,
				Messages: []model.Message{
					{
						ID:        "Greetings",
						Message:   "{Bonjour}",
						Positions: []string{"examples/simple/example.clj:10"},
					},
					{
						ID:        "Please confirm your email",
						Message:   "{Veuillez confirmer votre email}",
						Positions: []string{"examples/simple/example.clj:20"},
					},
					{
						ID:        "Welcome, %s!",
						Message:   "{Bienvenue, %s!}",
						Positions: []string{"examples/simple/example.clj:30"},
					},
					{
						ID:       "product",
						PluralID: "%s products",
						Message: `match {$count :number}
when 1 {produit}
when * {%s produits}
`,
						Positions: []string{
							"examples/simple/example.clj:40",
							"examples/simple/example.clj:50",
						},
					},
				},
			},
		},
		{
			name: "plural msgstr with simple msgstr",
			input: []byte(`msgid ""
							msgstr ""
							"Language: en\n"
							"Plural-Forms: nplurals=2; plural=(n != 1);\n"
							#. apple counts
							msgid "There is %d apple."
							msgid_plural "There are %d apples."
							msgstr[0] "Il y a %d pomme."
							msgstr[1] "Il y a %d pommes."

							msgid "hi"
							msgstr "ciao"
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
					{
						ID:      "hi",
						Message: "{ciao}",
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
			expectedErr: fmt.Errorf("convert tokens to pot.Po: get previous token: no previous token"),
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

	n := gofakeit.IntRange(1, 5)

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
