package pot

import (
	"fmt"
	"testing"

	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
)

func Test_TokensToPo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expected    Po
		expectedErr error
		input       []Token
	}{
		{
			name: "When all possible token values are provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "translator comment", Type: TranslatorComment},
				{Value: "extracted comment", Type: ExtractedComment},
				{Value: "reference", Type: Reference},
				{Value: "fuzzy", Type: Flag},
				{Value: "context", Type: MsgCtxt},
				{Value: "previous context", Type: MsgctxtPreviousContext},
				{Value: "msgid prev untranslated string", Type: MsgidPrevUntdStr},
				{Value: "There is 1 apple", Type: MsgId},
				{Value: "msgid plural prev untranslated string", Type: MsgidPluralPrevUntStrPlural},
				{Value: "There is %d apples", Type: PluralMsgId},
				{Value: "Il y a 1 pomme", Type: PluralMsgStr, Index: 0},
				{Value: "Il y a %d pommes", Type: PluralMsgStr, Index: 1},
				{Value: "message id", Type: MsgId},
				{Value: "message", Type: MsgStr},
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "Translator: John Doe",
					PluralForms: PluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						TranslatorComment:     "translator comment",
						ExtractedComment:      "extracted comment",
						Reference:             "reference",
						Flag:                  "fuzzy",
						MsgCtxt:               "context",
						MsgCtxtPrevCtxt:       "previous context",
						MsgIdPrevUnt:          "msgid prev untranslated string",
						MsgId:                 "There is 1 apple",
						MsgIdPrevUntPluralStr: "msgid plural prev untranslated string",
						MsgIdPlural:           "There is %d apples",
						MsgStr:                []string{"Il y a 1 pomme", "Il y a %d pommes"},
					},
					{
						MsgId:  "message id",
						MsgStr: []string{"message"},
					},
				},
			},
		},
		{
			name: "When msgid and msgstr token values are provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "message id", Type: MsgId},
				{Value: "message", Type: MsgStr},
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "Translator: John Doe",
					PluralForms: PluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						MsgId:  "message id",
						MsgStr: []string{"message"},
					},
				},
			},
		},
		{
			name: "When plural msgid and plural msgstr token values are provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "There is 1 apple", Type: MsgId},
				{Value: "There is %d apples", Type: PluralMsgId},
				{Value: "Il y a 1 pomme", Type: PluralMsgStr, Index: 0},
				{Value: "Il y a %d pommes", Type: PluralMsgStr, Index: 1},
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "Translator: John Doe",
					PluralForms: PluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						MsgId:       "There is 1 apple",
						MsgIdPlural: "There is %d apples",
						MsgStr:      []string{"Il y a 1 pomme", "Il y a %d pommes"},
					},
				},
			},
		},
		{
			name: "Invalid plural forms format is provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2", Type: HeaderPluralForms},
				{Value: "There is 1 apple", Type: MsgId},
				{Value: "There is %d apples", Type: PluralMsgId},
				{Value: "Il y a 1 pomme", Type: PluralMsgStr, Index: 0},
				{Value: "Il y a %d pommes", Type: PluralMsgStr, Index: 1},
			},
			expectedErr: fmt.Errorf("invalid plural forms format"),
		},
		{
			name: "Invalid nplurals value is provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=part; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "There is 1 apple", Type: MsgId},
				{Value: "There is %d apples", Type: PluralMsgId},
				{Value: "Il y a 1 pomme", Type: PluralMsgStr, Index: 0},
				{Value: "Il y a %d pommes", Type: PluralMsgStr, Index: 1},
			},
			expectedErr: fmt.Errorf("invalid nplurals value"),
		},
		{
			name: "Invalid nplurals part is provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: ; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "There is 1 apple", Type: MsgId},
				{Value: "There is %d apples", Type: PluralMsgId},
				{Value: "Il y a 1 pomme", Type: PluralMsgStr, Index: 0},
				{Value: "Il y a %d pommes", Type: PluralMsgStr, Index: 1},
			},
			expectedErr: fmt.Errorf("invalid nplurals part"),
		},
		{
			name: "Invalid plural string order is provided",
			input: []Token{
				{Value: "Language: en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "There is 1 apple", Type: MsgId},
				{Value: "There is %d apples", Type: PluralMsgId},
				{Value: "Il y a 1 pomme", Type: PluralMsgStr, Index: 1},
				{Value: "Il y a %d pommes", Type: PluralMsgStr, Index: 0},
			},
			expectedErr: fmt.Errorf("invalid plural string order %d", 1),
		},
		{
			name: "Invalid language header format is provided",
			input: []Token{
				{Value: "Language en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
				{Value: "message id", Type: MsgId},
				{Value: "message", Type: MsgStr},
			},
			expectedErr: fmt.Errorf("invalid token type %d", 0),
		},
		{
			name: "Invalid po file: no messages found",
			input: []Token{
				{Value: "Language en-US", Type: HeaderLanguage},
				{Value: "Translator: John Doe", Type: HeaderTranslator},
				{Value: "Plural-Forms: nplurals=2; plural=(n != 1);", Type: HeaderPluralForms},
			},
			expectedErr: fmt.Errorf("invalid po file: no messages found"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, _ := TokensToPo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
