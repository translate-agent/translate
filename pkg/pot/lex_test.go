package pot

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedErr error
		expected    []Token
	}{
		{
			name: "When all values are provided",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Translator: John Doe <johndoe@example.com> \\n\"\n" +
				"\"Language: en-US \\n\"\n" +
				"\"Plural-Forms: nplurals=2; plural=(n != 1); \\n\"\n" +
				"msgctxt \"ctxt\"\n" +
				"msgid \"id\"\n" +
				"msgstr \"str\"\n" +
				"msgid_plural \"There are %d oranges\"\n" +
				"msgstr[0] \"There is %d orange\"\n" +
				"msgstr[1] \"There are %d oranges\"\n" +
				"# translator-comment\n" +
				"#. extracted comment\n" +
				"#: reference\n" +
				"#, flag\n" +
				"#| msgctxt previous context\n" +
				"#| msgid previous id\n" +
				"#| msgid_plural previous id plural\n",
			expected: []Token{
				{Value: "", Type: MsgId},
				{Value: "", Type: MsgStr},
				{Value: "John Doe <johndoe@example.com>", Type: HeaderTranslator},
				{Value: "en-US", Type: HeaderLanguage},
				{Value: "nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "ctxt", Type: MsgCtxt},
				{Value: "id", Type: MsgId},
				{Value: "str", Type: MsgStr},
				{Value: "There are %d oranges", Type: PluralMsgId},
				{Value: "There is %d orange", Type: PluralMsgStr, Index: 0},
				{Value: "There are %d oranges", Type: PluralMsgStr, Index: 1},
				{Value: "translator-comment", Type: TranslatorComment},
				{Value: "extracted comment", Type: ExtractedComment},
				{Value: "reference", Type: Reference},
				{Value: "flag", Type: Flag},
				{Value: "msgctxt previous context", Type: MsgctxtPreviousContext},
				{Value: "msgid previous id", Type: MsgidPrevUntStr},
				{Value: "msgid_plural previous id plural", Type: MsgidPluralPrevUntStrPlural},
			},
		},
		{
			name: "When msgid and msgstr values are multiline",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Language: en-US\\n\"\n" +
				"msgid \"\"\n\"multiline id\"\n\"multiline id 2\"\n" +
				"msgstr \"\"\n\"text line 1\"\n\"next line 2\"\n",
			expected: []Token{
				{Value: "", Type: MsgId},
				{Value: "", Type: MsgStr},
				{Value: "en-US", Type: HeaderLanguage},
				{Value: "multiline id multiline id 2", Type: MsgId},
				{Value: "text line 1 next line 2", Type: MsgStr},
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
				{Value: "", Type: MsgId},
				{Value: "", Type: MsgStr},
				{Value: "en-US", Type: HeaderLanguage},
				{Value: "nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "multiline id multiline id 2", Type: MsgId},
				{Value: "There are %d oranges There are 1900000 oranges", Type: PluralMsgId},
				{Value: "There is %d orange There is 1 orange", Type: PluralMsgStr, Index: 0},
				{Value: "There are %d oranges There are 1900000 oranges", Type: PluralMsgStr, Index: 1},
			},
		},
		{
			name: "header Test",
			input: "msgid \"\"\n" +
				"msgstr \"\"\n" +
				"\"Language: en-US\\n\"\n" +
				"\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"\n",
			expected: []Token{
				{Value: "", Type: MsgId},
				{Value: "", Type: MsgStr},
				{Value: "en-US", Type: HeaderLanguage},
				{Value: "nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
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
				{Value: "", Type: MsgId},
				{Value: "", Type: MsgStr},
				{Value: "en-US", Type: HeaderLanguage},
				{Value: "nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "\"quoted\" id", Type: MsgId},
				{Value: "\"quoted\" str", Type: MsgStr},
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
