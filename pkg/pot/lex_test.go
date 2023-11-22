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
				mkToken(TokenTypeTranslatorComment, "Licensed to..."),
				mkToken(TokenTypeTranslatorComment, ""),
				mkToken(TokenTypeTranslatorComment, "http://www.apache.org/licenses/LICENSE-2.0"),
			},
		},
		{
			name: "When all values are provided",
			input: `msgid ""
msgstr ""
"Translator: John Doe <johndoe@example.com>\n"
"Language: en-US\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
"Project-Id-Version: 1.2\n"
"POT-Creation-Date: 10.02.2022.\n"
"PO-Revision-Date: 10.02.2022.\n"
"Last-Translator: John Doe\n"
"Language-Team: team\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"X-Generator: Poedit 2.2\n"
"Report-Msgid-Bugs-To: support@lingohub.com\n"
msgctxt "ctxt"
msgid "id"
msgstr "str"
msgid_plural "There are %d oranges"
msgstr[0] "There is %d orange"
msgstr[1] "There are %d oranges"
# translator-comment
#. extracted comment
#: reference1
#: reference2
#: reference3
#, flag
#| msgctxt previous context
#| msgid previous id
#| msgid_plural previous id plural
`,
			expected: []Token{
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderTranslator, "John Doe <johndoe@example.com>\\n"),
				mkToken(TokenTypeHeaderLanguage, "en-US\\n"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);\\n"),
				mkToken(TokenTypeHeaderProjectIDVersion, "1.2\\n"),
				mkToken(TokenTypeHeaderPOTCreationDate, "10.02.2022.\\n"),
				mkToken(TokenTypeHeaderPORevisionDate, "10.02.2022.\\n"),
				mkToken(TokenTypeHeaderLastTranslator, "John Doe\\n"),
				mkToken(TokenTypeHeaderLanguageTeam, "team\\n"),
				mkToken(TokenTypeHeaderMIMEVersion, "1.0\\n"),
				mkToken(TokenTypeHeaderContentType, "text/plain; charset=UTF-8\\n"),
				mkToken(TokenTypeHeaderContentTransferEncoding, "8bit\\n"),
				mkToken(TokenTypeHeaderXGenerator, "Poedit 2.2\\n"),
				mkToken(TokenTypeHeaderReportMsgidBugsTo, "support@lingohub.com\\n"),
				mkToken(TokenTypeMsgCtxt, "ctxt"),
				mkToken(TokenTypeMsgID, "id"),
				mkToken(TokenTypeMsgStr, "str"),
				mkToken(TokenTypePluralMsgID, "There are %d oranges"),
				mkToken(TokenTypePluralMsgStr, "There is %d orange", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "There are %d oranges", withIndex(1)),
				mkToken(TokenTypeTranslatorComment, "translator-comment"),
				mkToken(TokenTypeExtractedComment, "extracted comment"),
				mkToken(TokenTypeReference, "reference1"),
				mkToken(TokenTypeReference, "reference2"),
				mkToken(TokenTypeReference, "reference3"),
				mkToken(TokenTypeFlag, "flag"),
				mkToken(TokenTypeMsgctxtPreviousContext, "msgctxt previous context"),
				mkToken(TokenTypeMsgidPrevUntStr, "msgid previous id"),
				mkToken(TokenTypeMsgidPluralPrevUntStrPlural, "msgid_plural previous id plural"),
			},
		},
		{
			name: "When msgid and msgstr values are multiline",
			input: `msgid ""
msgstr ""
"Language: en-GB\n"
msgid ""
"multiline id"
"multiline id 2"
msgstr ""
"text line 1"
"next line 2"
`,
			expected: []Token{
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderLanguage, "en-GB\\n"),
				mkToken(TokenTypeMsgID, "multiline id multiline id 2"),
				mkToken(TokenTypeMsgStr, "text line 1 next line 2"),
			},
		},
		{
			name: "When msgid plural and msgstr plural values are multiline",
			input: `msgid ""
msgstr ""
"Language: en-US\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
msgid ""
"multiline id"
"multiline id 2"
msgid_plural "There are %d oranges"
"There are 1900000 oranges"
msgstr[0] "There is %d orange"
"There is 1 orange"
msgstr[1] "There are %d oranges"
"There are 1900000 oranges"
			`,
			expected: []Token{
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderLanguage, "en-US\\n"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);\\n"),
				mkToken(TokenTypeMsgID, "multiline id multiline id 2"),
				mkToken(TokenTypePluralMsgID, "There are %d oranges There are 1900000 oranges"),
				mkToken(TokenTypePluralMsgStr, "There is %d orange There is 1 orange", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "There are %d oranges There are 1900000 oranges", withIndex(1)),
			},
		},
		{
			name: "header Test",
			input: `msgid ""
msgstr ""
"Language: en-US\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
`,
			expected: []Token{
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderLanguage, "en-US\\n"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);\\n"),
			},
		},
		{
			name: "When msgid and msgstr values are quoted",
			input: `msgid ""
msgstr ""
"Language: en-US\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
msgid "\"quoted\" id"
msgstr "\"quoted\" str"
`,
			expected: []Token{
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderLanguage, "en-US\\n"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);\\n"),
				mkToken(TokenTypeMsgID, "\\\"quoted\\\" id"),
				mkToken(TokenTypeMsgStr, "\\\"quoted\\\" str"),
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
