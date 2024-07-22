package po

import (
	"testing"

	"go.expect.digital/translate/pkg/testutil/expect"
)

func TestPo_Marshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		want  string
		input PO
	}{
		{
			name: "singular",
			input: PO{
				Headers: Headers{
					{Name: "Language", Value: ""},
					{Name: "Last-Translator", Value: "John Doe"},
				},
				Messages: []Message{
					{
						MsgID:             "Hello, world!",
						MsgStr:            []string{},
						Flags:             []string{"fuzzy"},
						ExtractedComments: []string{"A simple greeting"},
						References:        []string{"main.go:1"},
					},
				},
			},
			want: `msgid ""
msgstr ""
"Language: \n"
"Last-Translator: John Doe\n"

#. A simple greeting
#: main.go:1
#, fuzzy
msgid "Hello, world!"
msgstr ""
`,
		},
		{
			name: "plural",
			input: PO{
				Headers: Headers{
					{Name: "Language", Value: "lv"},
					{Name: "Last-Translator", Value: "John Doe"},
					{Name: "Plural-Forms", Value: "nplurals=2; n != 1"},
				},
				Messages: []Message{
					{
						MsgID:       "There is 1 apple",
						MsgIDPlural: "There is 2 apples",
						MsgStr:      []string{"Ir 1 ābols", "Ir 2 āboli"},
					},
				},
			},
			want: `msgid ""
msgstr ""
"Language: lv\n"
"Last-Translator: John Doe\n"
"Plural-Forms: nplurals=2; n != 1\n"

msgid "There is 1 apple"
msgid_plural "There is 2 apples"
msgstr[0] "Ir 1 ābols"
msgstr[1] "Ir 2 āboli"
`,
		},
		{
			name: "multiline",
			input: PO{
				Headers: Headers{
					{Name: "Language", Value: "lv"},
					{Name: "Last-Translator", Value: "John Doe"},
					{Name: "Plural-Forms", Value: "nplurals=2; n != 1"},
				},
				Messages: []Message{
					{
						MsgID:  "\nThere is apple",
						MsgStr: []string{"\nIr ābols"},
					},
					{
						MsgID:       "\nThere is 1 orange",
						MsgIDPlural: "\nThere is multiple oranges",
						MsgStr:      []string{"\nIr 1 apelsīns", "\nIr vairāki apelsīni"},
					},
				},
			},
			want: `msgid ""
msgstr ""
"Language: lv\n"
"Last-Translator: John Doe\n"
"Plural-Forms: nplurals=2; n != 1\n"

msgid ""
"There is apple"
msgstr ""
"Ir ābols"

msgid ""
"There is 1 orange"
msgid_plural ""
"There is multiple oranges"
msgstr[0] ""
"Ir 1 apelsīns"
msgstr[1] ""
"Ir vairāki apelsīni"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.input.Marshal()

			expect.Equal(t, tt.want, string(got))
		})
	}
}
