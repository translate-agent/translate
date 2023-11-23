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
		// positive tests
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
				mkToken(TokenTypeHeaderTranslator, "John Doe <johndoe@example.com>"),
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
				mkToken(TokenTypeHeaderProjectIDVersion, "1.2"),
				mkToken(TokenTypeHeaderPOTCreationDate, "10.02.2022."),
				mkToken(TokenTypeHeaderPORevisionDate, "10.02.2022."),
				mkToken(TokenTypeHeaderLastTranslator, "John Doe"),
				mkToken(TokenTypeHeaderLanguageTeam, "team"),
				mkToken(TokenTypeHeaderMIMEVersion, "1.0"),
				mkToken(TokenTypeHeaderContentType, "text/plain; charset=UTF-8"),
				mkToken(TokenTypeHeaderContentTransferEncoding, "8bit"),
				mkToken(TokenTypeHeaderXGenerator, "Poedit 2.2"),
				mkToken(TokenTypeHeaderReportMsgidBugsTo, "support@lingohub.com"),
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
				mkToken(TokenTypeHeaderLanguage, "en-GB"),
				mkToken(TokenTypeMsgID, "\nmultiline id\nmultiline id 2"),
				mkToken(TokenTypeMsgStr, "\ntext line 1\nnext line 2"),
			},
		},
		{
			name: "pot with plural and escaped newline",
			input: `msgid ""
msgstr ""
"Language: en-US\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid ""
"There is %d orange\n"
"that is on the tree"
msgid_plural ""
"There are %d oranges\n"
"that are on the tree"
msgstr[0] ""
"There is %d orange\n"
"that is on the tree"
msgstr[1] ""
"There are %d oranges\n"
"that are on the tree"
`,
			expected: []Token{
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
				mkToken(TokenTypeMsgID, "\nThere is %d orange\\n\nthat is on the tree"),
				mkToken(TokenTypePluralMsgID, "\nThere are %d oranges\\n\nthat are on the tree"),
				mkToken(TokenTypePluralMsgStr, "\nThere is %d orange\\n\nthat is on the tree", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "\nThere are %d oranges\\n\nthat are on the tree", withIndex(1)),
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
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
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
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
				mkToken(TokenTypeMsgID, "\\\"quoted\\\" id"),
				mkToken(TokenTypeMsgStr, "\\\"quoted\\\" str"),
			},
		},
		{
			name: "Multiline msgid with leading spaces",
			input: `msgid ""
"Add filter clauses to control the filter's source query,\n"
"                    though only in the context of the autocomplete i.e.,"
"these conditions\n"
"                    do not impact how the filter is applied to the"
"dashboard. This is useful\n"
"                    when you want to improve the query's performance by"
"only scanning a subset\n"
"                    of the underlying data or limit the available values"
"displayed in the filter."
`,
			expected: []Token{
				mkToken(TokenTypeMsgID, `
Add filter clauses to control the filter's source query,\n
                    though only in the context of the autocomplete i.e.,
these conditions\n
                    do not impact how the filter is applied to the
dashboard. This is useful\n
                    when you want to improve the query's performance by
only scanning a subset\n
                    of the underlying data or limit the available values
displayed in the filter.`),
			},
		},
		// negative tests
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
