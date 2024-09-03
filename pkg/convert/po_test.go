package convert

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// requireEqualPO is a helper function to compare two PO strings, ignoring whitespace, newlines, and quotes.
func requireEqualPO(t *testing.T, want, got string) {
	t.Helper()

	replace := func(s string) string { return strings.NewReplacer("\\n", "", "\n", "", "\"", "").Replace(s) }

	if replace(want) != replace(got) {
		t.Errorf("want po '%s', got '%s'", replace(want), replace(got))
	}
}

// Test_FromPoSingular tests the conversion from PO->Translation->PO for singular messages.
func Test_PoSingular(t *testing.T) {
	t.Parallel()

	type args struct {
		original *bool
		input    string
	}

	tests := []struct {
		name string
		args args
		want model.Translation
	}{
		// Without placeholders
		{
			name: "simple original",
			args: args{
				original: nil,
				input: `msgid "Hello, world!"
msgstr ""

msgid "Goodbye!"
msgstr ""
`,
			},
			want: model.Translation{
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
			want: model.Translation{
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
			want: model.Translation{
				Language: language.German,
				Original: false,
				Messages: []model.Message{
					{
						ID:      " Set the opacity to 0 if you do not want to override the color specified in the GeoJSON",
						Message: " Setzen Sie die Deckkraft auf 0, wenn Sie die im GeoJSON angegebene Farbe nicht überschreiben möchten.", //nolint:lll
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
			name: "original with placeholders",
			args: args{
				original: nil,
				input: `#, python-format
msgid "Hello, {name}!"
msgstr ""

#, c-format
msgid "%s, Hello!"
msgstr ""

#, python-format
msgid "Hello, %(name)s, world!"
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
			want: model.Translation{
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
						ID:     "%s, Hello!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { c-format }
.local $ph0 = { |%s| }
{{{ $ph0 }, Hello!}}`,
					},
					{
						ID:     "Hello, %(name)s, world!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { python-format }
.local $name = { |%(name)s| }
{{Hello, { $name }, world!}}`,
					},
					{
						ID:     "Hello, {}!",
						Status: model.MessageStatusTranslated,
						Message: `.local $format = { python-format }
.local $ph0 = { |{}| }
{{Hello, { $ph0 }!}}`,
					},
					{
						ID: "Hello, %s!",
						Message: `.local $format = { no-c-format }
{{Hello, %s!}}`,
						Status: model.MessageStatusTranslated,
					},
					{
						ID:      "Hello, %s!",
						Message: "Hello, %s!",
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "translation with placeholders",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""
"Language: lv\n"

#, python-format
msgid "Hello, {name}!"
msgstr "Sveika, {name}!"
`,
			},
			want: model.Translation{
				Language: language.Latvian,
				Original: false,
				Messages: []model.Message{
					{
						ID:     "Hello, {name}!",
						Status: model.MessageStatusUntranslated,
						Message: `.local $format = { python-format }
.local $name = { |{name}| }
{{Sveika, { $name }!}}`,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Test: PO -> Translation

			got, err := FromPo([]byte(test.args.input), test.args.original)
			if err != nil {
				t.Error(err)
				return
			}

			assertTranslation(t, test.want, got)

			// Test: Translation -> PO

			gotPo, err := ToPo(got)
			if err != nil {
				t.Error(err)
				return
			}

			requireEqualPO(t, test.args.input, string(gotPo))
		})
	}
}

// Test_FromPoPlural tests the conversion from PO->Translation->PO for plural messages.
func Test_PoPlural(t *testing.T) {
	t.Parallel()

	type args struct {
		original *bool
		input    string
	}

	tests := []struct {
		name string
		args args
		want model.Translation
	}{
		// Without placeholders
		{
			name: "simple original",
			args: args{
				original: nil,
				input: `#. Description
msgid "There is one apple."
msgid_plural "There are multiple apples."
msgstr[0] ""
msgstr[1] ""
`,
			},
			want: model.Translation{
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
			name: "simple translation",
			args: args{
				original: nil,
				input: `msgid ""
msgstr ""
"Language: ru\n"

#: superset-frontend/src/filters/components/GroupBy/GroupByFilterPlugin.tsx:87
msgid "option"
msgid_plural "options"
msgstr[0] "вариант"
msgstr[1] "варианта"
msgstr[2] "вариантов"
`,
			},
			want: model.Translation{
				Language: language.Russian,
				Original: false,
				Messages: []model.Message{
					{
						ID:       "option",
						PluralID: "options",
						Message: `.match { $count }
1 {{вариант}}
2 {{варианта}}
* {{вариантов}}`,
						Status: model.MessageStatusUntranslated,
						Positions: []string{
							"superset-frontend/src/filters/components/GroupBy/GroupByFilterPlugin.tsx:87",
						},
					},
				},
			},
		},
		// With placeholders
		{
			name: "original with placeholders",
			args: args{
				original: ptr(true),
				input: `#: superset-frontend/src/components/ErrorMessage/ParameterErrorMessage.tsx:88
#, python-format
msgid "%(suggestion)s instead of \"%(undefinedParameter)s?\""
msgid_plural ""
"%(firstSuggestions)s or %(lastSuggestion)s instead of"
"\"%(undefinedParameter)s\"?"
msgstr[0] ""
msgstr[1] ""
`,
			},
			want: model.Translation{
				Language: language.Und,
				Original: true,
				Messages: []model.Message{
					{
						ID:        "%(suggestion)s instead of \\\"%(undefinedParameter)s?\\\"",
						PluralID:  "%(firstSuggestions)s or %(lastSuggestion)s instead of\\\"%(undefinedParameter)s\\\"?",
						Status:    model.MessageStatusTranslated,
						Positions: []string{"superset-frontend/src/components/ErrorMessage/ParameterErrorMessage.tsx:88"},
						Message: `.local $format = { python-format }
.local $suggestion = { |%(suggestion)s| }
.local $undefinedParameter = { |%(undefinedParameter)s| }
.local $firstSuggestions = { |%(firstSuggestions)s| }
.local $lastSuggestion = { |%(lastSuggestion)s| }
.match { $count }
1 {{{ $suggestion } instead of \\"{ $undefinedParameter }?\\"}}
* {{{ $firstSuggestions } or { $lastSuggestion } instead of\\"{ $undefinedParameter }\\"?}}`,
					},
				},
			},
		},
		{
			name: "translation with placeholders",
			args: args{
				original: ptr(false),
				input: `msgid ""
msgstr ""
"Language: ru\n"

#: superset-frontend/src/components/ErrorMessage/ParameterErrorMessage.tsx:88
#, python-format
msgid "%(suggestion)s instead of \"%(undefinedParameter)s?\""
msgid_plural ""
"%(firstSuggestions)s or %(lastSuggestion)s instead of "
"\"%(undefinedParameter)s\"?"
msgstr[0] "%(suggestion)s вместо \"%(undefinedParameter)s\"?"
msgstr[1] ""
"%(firstSuggestions)s или %(lastSuggestion)s вместо "
"\"%(undefinedParameter)s\"?"
msgstr[2] ""
"%(firstSuggestions)s или %(lastSuggestion)s вместо "
"\"%(undefinedParameter)s\"?"
`,
			},
			want: model.Translation{
				Language: language.Russian,
				Original: false,
				Messages: []model.Message{
					{
						ID:        "%(suggestion)s instead of \\\"%(undefinedParameter)s?\\\"",
						PluralID:  "%(firstSuggestions)s or %(lastSuggestion)s instead of \\\"%(undefinedParameter)s\\\"?",
						Status:    model.MessageStatusUntranslated,
						Positions: []string{"superset-frontend/src/components/ErrorMessage/ParameterErrorMessage.tsx:88"},
						Message: `.local $format = { python-format }
.local $suggestion = { |%(suggestion)s| }
.local $undefinedParameter = { |%(undefinedParameter)s| }
.local $firstSuggestions = { |%(firstSuggestions)s| }
.local $lastSuggestion = { |%(lastSuggestion)s| }
.match { $count }
1 {{{ $suggestion } вместо \\"{ $undefinedParameter }\\"?}}
2 {{{ $firstSuggestions } или { $lastSuggestion } вместо \\"{ $undefinedParameter }\\"?}}
* {{{ $firstSuggestions } или { $lastSuggestion } вместо \\"{ $undefinedParameter }\\"?}}`,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Test: PO -> Translation

			got, err := FromPo([]byte(test.args.input), test.args.original)
			if err != nil {
				t.Log("\n" + test.args.input)
				t.Error(err)

				return
			}

			assertTranslation(t, test.want, got)

			// Test: Translation -> PO

			gotPo, err := ToPo(got)
			if err != nil {
				t.Error(err)
				return
			}

			requireEqualPO(t, test.args.input, string(gotPo))
		})
	}
}

func assertTranslation(t *testing.T, want, got model.Translation) {
	t.Helper()

	cmpTag := cmp.Comparer(func(a, b language.Tag) bool {
		return a == b
	})

	if v := cmp.Diff(want, got, cmpTag); v != "" {
		t.Errorf("want equal translations\n%s", v)
	}
}
