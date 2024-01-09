package convert

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

// -–––--------------------------PO->Translation-----------------------–––-----

func Test_FromPoSingular(t *testing.T) {
	t.Parallel()

	type args struct {
		original *bool
		input    string
	}

	tests := []struct {
		name     string
		args     args
		expected model.Translation
	}{
		// Without placeholders
		{
			name: "simple original",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""
"Language: en\n"

msgid "Hello, world!"
msgstr ""

msgid "Goodbye!"
msgstr ""
`,
			},
			expected: model.Translation{
				Language: language.English,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: "Hello, world!",
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "Goodbye!",
						Message: "Goodbye!",
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "simple translation",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""
"Language: lv\n"

msgid "Hello, world!"
msgstr "Sveika, pasaule!"

msgid "Goodbye!"
msgstr ""

#, fuzzy
msgid "Dinosaurs"
msgstr "Dinozauri"
`,
			},
			expected: model.Translation{
				Language: language.Latvian,
				Original: false,
				Messages: []model.Message{
					{
						ID:      "Hello, world!",
						Message: "Sveika, pasaule!",
						Status:  model.MessageStatusUntranslated,
					},
					{
						ID:      "Goodbye!",
						Message: "",
						Status:  model.MessageStatusUntranslated,
					},
					{
						ID:      "Dinosaurs",
						Message: "Dinozauri",
						Status:  model.MessageStatusFuzzy,
					},
				},
			},
		},
		{
			name: "multiline translation",
			args: args{
				original: ptr(false),
				input: `msgid ""
msgstr ""
"Language: de\n"

#: superset-frontend/plugins/legacy-preset-chart-deckgl/src/utilities/Shared_DeckGL.jsx:222
#: superset-frontend/plugins/legacy-preset-chart-deckgl/src/utilities/Shared_DeckGL.jsx:235
msgid ""
" Set the opacity to 0 if you do not want to override the color specified "
"in the GeoJSON"
msgstr ""
" Setzen Sie die Deckkraft auf 0, wenn Sie die im GeoJSON angegebene Farbe"
" nicht überschreiben möchten."
`,
			},
			expected: model.Translation{
				Language: language.German,
				Original: false,
				Messages: []model.Message{
					{
						ID:      "\n Set the opacity to 0 if you do not want to override the color specified \nin the GeoJSON",
						Message: "\n Setzen Sie die Deckkraft auf 0, wenn Sie die im GeoJSON angegebene Farbe\n nicht überschreiben möchten.", //nolint:lll
						Status:  model.MessageStatusUntranslated,
						Positions: []string{
							"superset-frontend/plugins/legacy-preset-chart-deckgl/src/utilities/Shared_DeckGL.jsx:222",
							"superset-frontend/plugins/legacy-preset-chart-deckgl/src/utilities/Shared_DeckGL.jsx:235",
						},
					},
				},
			},
		},
		// With placeholders
		{
			name: "placeholders",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""

#, python-format
msgid "Hello, {name}!"
msgstr ""

#, c-format
msgid "Hello, %s!"
msgstr ""

#, python-format
msgid "Hello, %(name)s!"
msgstr ""

#, python-format
msgid "Hello, {}!"
msgstr ""

#, no-c-format
msgid "Hello, %s!"
msgstr ""

msgid "Hello, %s!"
msgstr ""
`,
			},
			expected: model.Translation{
				Language: language.Und,
				Original: true,
				Messages: []model.Message{
					{
						ID:     "Hello, {name}!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { python-format }
.local $name = { |{name}| }
{{Hello, { $name }!}}`,
					},
					{
						ID:     "Hello, %s!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { c-format }
.local $ph0 = { |%s| }
{{Hello, { $ph0 }!}}`,
					},
					{
						ID:     "Hello, %(name)s!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { python-format }
.local $name = { |%(name)s| }
{{Hello, { $name }!}}`,
					},
					{
						ID:     "Hello, {}!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { python-format }
.local $ph0 = { |{}| }
{{Hello, { $ph0 }!}}`,
					},
					{
						ID:      "Hello, %s!",
						Message: "Hello, %s!",
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "Hello, %s!",
						Message: "Hello, %s!",
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromPo([]byte(tt.args.input), tt.args.original)
			require.NoError(t, err)

			testutil.EqualTranslations(t, &tt.expected, &actual)
		})
	}
}

func Test_FromPoPlural(t *testing.T) {
	t.Parallel()

	type args struct {
		original *bool
		input    string
	}

	tests := []struct {
		name     string
		args     args
		expected model.Translation
	}{
		// Without placeholders
		{
			name: "original",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""

#. Description
msgid "There is one apple."
msgid_plural "There are multiple apples."
msgstr[0] ""
msgstr[1] ""
`,
			},
			expected: model.Translation{
				Language: language.Und,
				Original: true,
				Messages: []model.Message{
					{
						ID:          "There is one apple.",
						PluralID:    "There are multiple apples.",
						Message:     ".match { $count }\n1 {{There is one apple.}}\n* {{There are multiple apples.}}",
						Description: "Description",
						Status:      model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "translation",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""
"Language: ru\n"
"Plural-Forms: nplurals=3; plural=(n%10==1 && n%100!=11 ? 0 : n%10>=2 && "
"n%10<=4 && (n%100<12 || n%100>14) ? 1 : 2)\n"

#: superset-frontend/src/filters/components/GroupBy/GroupByFilterPlugin.tsx:87
msgid "option"
msgid_plural "options"
msgstr[0] "вариант"
msgstr[1] "варианта"
msgstr[2] "вариантов"
`,
			},
			expected: model.Translation{
				Language: language.Russian,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "option",
						PluralID: "options",
						Message:  ".match { $count }\n1 {{вариант}}\n2 {{варианта}}\n* {{вариантов}}",
						Status:   model.MessageStatusUntranslated,
						Positions: []string{
							"superset-frontend/src/filters/components/GroupBy/GroupByFilterPlugin.tsx:87",
						},
					},
				},
			},
		},
		// With placeholders
		{
			name: "placeholders",
			args: args{
				original: ptr(true),
				input: `msgid ""
msgstr ""

#: superset-frontend/src/components/ErrorMessage/ParameterErrorMessage.tsx:88
#, python-format
msgid "%(suggestion)s instead of \"%(undefinedParameter)s?\""
msgid_plural ""
"%(firstSuggestions)s or %(lastSuggestion)s instead of"
"\"%(undefinedParameter)s\"?"
msgstr[0] ""
msgstr[1] ""
`,
			},
			expected: model.Translation{
				Language: language.Und,
				Original: true,
				Messages: []model.Message{
					{
						ID:        "%(suggestion)s instead of \\\"%(undefinedParameter)s?\\\"",
						PluralID:  "\n%(firstSuggestions)s or %(lastSuggestion)s instead of\n\\\"%(undefinedParameter)s\\\"?",
						Status:    model.MessageStatusTranslated,
						Positions: []string{"superset-frontend/src/components/ErrorMessage/ParameterErrorMessage.tsx:88"},
						Message: `.local $format = { python-format }
.local $suggestion = { |%(suggestion)s| }
.local $undefinedParameter = { |%(undefinedParameter)s| }
.local $firstSuggestions = { |%(firstSuggestions)s| }
.local $lastSuggestion = { |%(lastSuggestion)s| }
.match { $count }
1 {{{ $suggestion }  instead of \\" { $undefinedParameter } ?\\"}}
* {{
 { $firstSuggestions }  or  { $lastSuggestion }  instead of
\\" { $undefinedParameter } \\"?}}`,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromPo([]byte(tt.args.input), tt.args.original)
			require.NoError(t, err)

			testutil.EqualTranslations(t, &tt.expected, &actual)
		})
	}
}

// -–––--------------------------Translation->PO-----------------------–––-----

func Test_ToPot(t *testing.T) {
	t.Parallel()

	t.Skip() // TODO

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

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

			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_TransformMessage(t *testing.T) {
	t.Skip() // TODO
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

	restoredTranslation, err := FromPo(msgPot, &translation.Original)
	require.NoError(t, err)

	assert.Equal(t, translation, restoredTranslation)
}
