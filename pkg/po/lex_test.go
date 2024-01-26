package po

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
"Generated-By: Babel 2.9.1\n"
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
				mkToken(TokenTypeHeaderGeneratedBy, "Babel 2.9.1"),
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
			name: "po with plural and escaped newline",
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
		{
			name: "Multiline msgid escaped quote at the end",
			input: `msgid ""
"If duplicate columns are not overridden, they will be presented as \"X.1,"
" X.2 ...X.x\""`,
			expected: []Token{
				mkToken(TokenTypeMsgID, "\nIf duplicate columns are not overridden, they will be presented as \\\"X.1,\n X.2 ...X.x\\\""), //nolint:lll
			},
		},
		{
			name:     "language header, empty value",
			input:    `"Language:"`,
			expected: []Token{mkToken(TokenTypeHeaderLanguage, "")},
		},
		{
			name:     "language header, without whitespace separator",
			input:    `"Language:en-US"`,
			expected: []Token{mkToken(TokenTypeHeaderLanguage, "en-US")},
		},
		{
			name:     "language header, value enclosed in spaces",
			input:    `"Language:   en-US"  `,
			expected: []Token{mkToken(TokenTypeHeaderLanguage, "en-US")},
		},
		{
			name:     "message id, value enclosed in spaces",
			input:    `msgid    "quoted"   `,
			expected: []Token{mkToken(TokenTypeMsgID, "quoted")},
		},
		{
			name: "message, multiline",
			input: `msgstr "hello "   
			"world"`,
			expected: []Token{mkToken(TokenTypeMsgStr, "hello \nworld")},
		},
		{
			name:     "plural message, value enclosed in spaces",
			input:    `msgstr[0]   "message"   `,
			expected: []Token{mkToken(TokenTypePluralMsgStr, "message")},
		},
		{
			name:     "translator comment, empty value",
			input:    `#`,
			expected: []Token{mkToken(TokenTypeTranslatorComment, "")},
		},
		{
			name:     "msgctxt comment",
			input:    `#| msgctxt context`,
			expected: []Token{mkToken(TokenTypeMsgctxtPreviousContext, "msgctxt context")},
		},
		{
			name:     "msgctxt comment, value enclosed in spaces",
			input:    `#| msgctxt     context    `,
			expected: []Token{mkToken(TokenTypeMsgctxtPreviousContext, "msgctxt context")},
		},
		// negative tests
		{
			name:        "language header, missing closing quotes",
			input:       `"Language: en-US`,
			expectedErr: fmt.Errorf("line must end with double quotation mark - \" "),
		},
		{
			name: "language header, missing opening quotes",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"Language: en-US\n" +
				"Plural-Forms: nplurals=2; plural=(n != 1);\n" +
				"msgid\"id\"\n" +
				"msgstr \"\"quoted\" str\"\n",
			expectedErr: fmt.Errorf("unknown line prefix"),
		},
		{
			name:        "message id, missing whitespace separator",
			input:       `msgid"quoted"`,
			expectedErr: fmt.Errorf("value must be prefixed with space"),
		},
		{
			name:        "message id, missing closing quotes",
			input:       `msgid "\"quoted\" id`,
			expectedErr: fmt.Errorf("value must be enclosed in double quotation mark - \"\" "),
		},
		{
			name: "message, multiline missing closing double quotation mark",
			input: `msgstr "hello "   
			"world`,
			expectedErr: fmt.Errorf("line must end with double quotation mark - \" "),
		},
		{
			name:        "plural message, invalid index format",
			input:       `msgstr[-1]`,
			expectedErr: fmt.Errorf("invalid syntax"),
		},
		{
			name:        "msgctxt comment, missing whitespace separator",
			input:       `#| msgctxtcontext`,
			expectedErr: fmt.Errorf("value must be prefixed with space"),
		},
		{
			name:        "translator comment, missing whitespace separator",
			input:       `#comment`,
			expectedErr: fmt.Errorf("value must be prefixed with space"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			result, err := lex(r)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}