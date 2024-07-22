package po

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/testutil/expect"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  PO
	}{
		{
			name: "only headers",
			input: `# Top-level comment1
# Top-level comment2
msgid ""
msgstr ""
"Project-Id-Version: Hello World 1.0\n"`,
			want: PO{
				Headers: Headers{
					{Name: "Project-Id-Version", Value: "Hello World 1.0"},
				},
			},
		},
		{
			name: "only messages",
			input: `#, fuzzy
#: main.go:1
# Translator comment
msgid "id1"
msgstr "str1"

#. Extracted comment
msgid "id2"
msgid_plural "id2 plural"
msgstr[0] "str2"
msgstr[1] "str2-1"`,
			want: PO{
				Messages: []Message{
					{
						MsgID:              "id1",
						MsgStr:             []string{"str1"},
						Flags:              []string{"fuzzy"},
						References:         []string{"main.go:1"},
						TranslatorComments: []string{"Translator comment"},
					},
					{
						MsgID:             "id2",
						MsgIDPlural:       "id2 plural",
						MsgStr:            []string{"str2", "str2-1"},
						ExtractedComments: []string{"Extracted comment"},
					},
				},
			},
		},
		{
			name: "full example",
			input: `# Top-level comment1
# Top-level comment2
msgid ""
msgstr ""
"Project-Id-Version: Hello World 1.0\n"
"Report-Msgid-Bugs-To: \n"
"POT-Creation-Date: 2023-05-16 13:48+0000\n"
"PO-Revision-Date: 2022-09-22 05:46+0000\n"
"Last-Translator: Jane Doe, 2023\n"
"Language-Team: Latvian\n"
"Language: lv\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: \n"
"Plural-Forms: nplurals=3; plural=n%10==1 && n%100!=11 ? 0 : n%10>=2 && n"
"%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2;\n"

msgid "id1"
msgstr "str1"

msgid ""
"multiline id1"
msgstr ""
"multiline str1"

msgid "id2"
msgid_plural "id2 plural"
msgstr[0] "str2"
msgstr[1] "str2-1"
msgstr[2] "str2-2"

msgid ""
"multiline \n"
"plural"
msgid_plural ""
"multiline \n"
"plurals"
msgstr[0] ""
"str3"
msgstr[1] ""
"str3-1"
msgstr[2] ""
"str3-2"

# Translator comment
#. Extracted comment
#: main.go:1
#, flag
msgid "Hello, world!"
msgstr "Hello, world!"`,
			want: PO{
				Headers: Headers{
					{Name: "Project-Id-Version", Value: "Hello World 1.0"},
					{Name: "Report-Msgid-Bugs-To", Value: ""},
					{Name: "POT-Creation-Date", Value: "2023-05-16 13:48+0000"},
					{Name: "PO-Revision-Date", Value: "2022-09-22 05:46+0000"},
					{Name: "Last-Translator", Value: "Jane Doe, 2023"},
					{Name: "Language-Team", Value: "Latvian"},
					{Name: "Language", Value: "lv"},
					{Name: "MIME-Version", Value: "1.0"},
					{Name: "Content-Type", Value: "text/plain; charset=UTF-8"},
					{Name: "Content-Transfer-Encoding", Value: ""},
					{Name: "Plural-Forms", Value: `nplurals=3; plural=n%10==1 && n%100!=11 ? 0 : n%10>=2 && n
%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2;`},
				},
				Messages: []Message{
					{
						MsgID:  "id1",
						MsgStr: []string{"str1"},
					},
					{
						MsgID:  "multiline id1",
						MsgStr: []string{"multiline str1"},
					},
					{
						MsgID:       "id2",
						MsgIDPlural: "id2 plural",
						MsgStr:      []string{"str2", "str2-1", "str2-2"},
					},
					{
						MsgID:       "multiline \nplural",
						MsgIDPlural: "multiline \nplurals",
						MsgStr:      []string{"str3", "str3-1", "str3-2"},
					},
					{
						MsgID:              "Hello, world!",
						MsgStr:             []string{"Hello, world!"},
						TranslatorComments: []string{"Translator comment"},
						ExtractedComments:  []string{"Extracted comment"},
						References:         []string{"main.go:1"},
						Flags:              []string{"flag"},
					},
				},
			},
		},
		{
			name: "multiple lines",
			input: `#: superset-frontend/src/explore/components/controls/DndColumnSelectControl/Option.tsx:71
#: superset-frontend/src/explore/components/controls/OptionControls/index.tsx:326
msgid ""
"\n"
"                This filter was inherited from the dashboard's context.\n"
"                It won't be saved when saving the chart.\n"
"              "
msgstr ""`,
			want: PO{
				Messages: []Message{
					{
						MsgID: `
                This filter was inherited from the dashboard's context.
                It won't be saved when saving the chart.
              `,
						MsgStr: []string{},
						References: []string{
							"superset-frontend/src/explore/components/controls/DndColumnSelectControl/Option.tsx:71",
							"superset-frontend/src/explore/components/controls/OptionControls/index.tsx:326",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse([]byte(tt.input))
			expect.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
