package pot

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedErr error
		expected    []Token
	}{
		// TODO(jhorsts): this is not actually translator comment
		// but file level comment. Should it be simply TypeComment?
		// https://raw.githubusercontent.com/apache/superset/master/superset/translations/messages.pot
		{
			name: "multiline comment",
			input: "# Licensed to...\n" +
				"#\n" +
				"# http://www.apache.org/licenses/LICENSE-2.0",
			expected: []Token{
				tokenTranslatorComment("Licensed to..."),
				tokenTranslatorComment(""),
				tokenTranslatorComment("http://www.apache.org/licenses/LICENSE-2.0"),
			},
		},
		{
			name: "When all values are provided",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Translator: John Doe <johndoe@example.com>\\n\"\n" +
				"\"Language: en-US\\n\"\n" +
				"\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"\n" +
				"\"Project-Id-Version: 1.2\\n\"\n" +
				"\"POT-Creation-Date: 10.02.2022.\\n\"\n" +
				"\"PO-Revision-Date: 10.02.2022.\\n\"\n" +
				"\"Last-Translator: John Doe\\n\"\n" +
				"\"Language-Team: team\\n\"\n" +
				"\"MIME-Version: 1.0\\n\"\n" +
				"\"Content-Type: text/plain; charset=UTF-8\\n\"\n" +
				"\"Content-Transfer-Encoding: 8bit\\n\"\n" +
				"\"X-Generator: Poedit 2.2\\n\"\n" +
				"\"Report-Msgid-Bugs-To: support@lingohub.com\\n\"\n" +
				"msgctxt \"ctxt\"\n" +
				"msgid \"id\"\n" +
				"msgstr \"str\"\n" +
				"msgid_plural \"There are %d oranges\"\n" +
				"msgstr[0] \"There is %d orange\"\n" +
				"msgstr[1] \"There are %d oranges\"\n" +
				"# translator-comment\n" +
				"#. extracted comment\n" +
				"#: reference1\n" +
				"#: reference2\n" +
				"#: reference3\n" +
				"#, flag\n" +
				"#| msgctxt previous context\n" +
				"#| msgid previous id\n" +
				"#| msgid_plural previous id plural\n",
			expected: []Token{
				tokenMsgID(""),
				tokenMsgStr(""),
				tokenHeaderTranslator("John Doe <johndoe@example.com>"),
				tokenHeaderLanguage("en-US"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenHeaderProjectIDVersion("1.2"),
				tokenHeaderPOTCreationDate("10.02.2022."),
				tokenHeaderPORevisionDate("10.02.2022."),
				tokenHeaderLastTranslator("John Doe"),
				tokenHeaderLanguageTeam("team"),
				tokenHeaderMIMEVersion("1.0"),
				tokenHeaderContentType("text/plain; charset=UTF-8"),
				tokenHeaderContentTransferEncoding("8bit"),
				tokenHeaderXGenerator("Poedit 2.2"),
				tokenHeaderReportMsgidBugsTo("support@lingohub.com"),
				tokenMsgCtxt("ctxt"),
				tokenMsgID("id"),
				tokenMsgStr("str"),
				tokenPluralMsgID("There are %d oranges"),
				tokenPluralMsgStr("There is %d orange", 0),
				tokenPluralMsgStr("There are %d oranges", 1),
				tokenTranslatorComment("translator-comment"),
				tokenExtractedComment("extracted comment"),
				tokenReference("reference1"),
				tokenReference("reference2"),
				tokenReference("reference3"),
				tokenFlag("flag"),
				tokenMsgctxtPreviousContext("msgctxt previous context"),
				tokenMsgidPrevUntStr("msgid previous id"),
				tokenMsgidPluralPrevUntStrPlural("msgid_plural previous id plural"),
			},
		},
		{
			name: "When msgid and msgstr values are multiline",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Language: en-GB\\n\"\n" +
				"msgid \"\"\n\"multiline id\"\n\"multiline id 2\"\n" +
				"msgstr \"\"\n\"text line 1\"\n\"next line 2\"\n",
			expected: []Token{
				tokenMsgID(""),
				tokenMsgStr(""),
				tokenHeaderLanguage("en-GB"),
				tokenMsgID("multiline id multiline id 2"),
				tokenMsgStr("text line 1 next line 2"),
			},
		},
		{
			name: "When msgid plural and msgstr plural values are multiline",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Language: en-US\\n\"\n" +
				"\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"\n" +
				"msgid \"\"\n\"multiline id\"\n\"multiline id 2\"\n" +
				"msgid_plural \"There are %d oranges\"\n\"There are 1900000 oranges\"\n" +
				"msgstr[0] \"There is %d orange\"\n\"There is 1 orange\"\n" +
				"msgstr[1] \"There are %d oranges\"\n\"There are 1900000 oranges\"\n",
			expected: []Token{
				tokenMsgID(""),
				tokenMsgStr(""),
				tokenHeaderLanguage("en-US"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenMsgID("multiline id multiline id 2"),
				tokenPluralMsgID("There are %d oranges There are 1900000 oranges"),
				tokenPluralMsgStr("There is %d orange There is 1 orange", 0),
				tokenPluralMsgStr("There are %d oranges There are 1900000 oranges", 1),
			},
		},
		{
			name: "header Test",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Language: en-US\\n\"\n" +
				"\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"\n",
			expected: []Token{
				tokenMsgID(""),
				tokenMsgStr(""),
				tokenHeaderLanguage("en-US"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
			},
		},
		{
			name: "When msgid and msgstr values are quoted",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Language: en-US\\n\"\n" +
				"\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"\n" +
				"msgid \"\"quoted\" id\"\n" +
				"msgstr \"\"quoted\" str\"\n",
			expected: []Token{
				tokenMsgID(""),
				tokenMsgStr(""),
				tokenHeaderLanguage("en-US"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenMsgID("\"quoted\" id"),
				tokenMsgStr("\"quoted\" str"),
			},
		},
		{
			name: "When msgid value is incorrect",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"Language: en-US\n" +
				"Plural-Forms: nplurals=2; plural=(n != 1);\n" +
				"msgid\"id\"\n" +
				"msgstr \"\"quoted\" str\"\n",
			expectedErr: fmt.Errorf("incorrect format of msgid: incorrect format of po tags"),
		},
		{
			name: "When msgstr value is incorrect",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"Language: en-US\n" +
				"Plural-Forms: nplurals=2; plural=(n != 1);\n" +
				"msgid \"id\"\n" +
				"msgstr\"str\"\n",
			expectedErr: fmt.Errorf("incorrect format of msgstr: incorrect format of po tags"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			result, err := Lex(r)

			if tt.expectedErr != nil {
				require.Errorf(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}
