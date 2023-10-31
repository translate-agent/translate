package pot

import (
	"fmt"
	"testing"

	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				tokenMsgID(""),
				tokenMsgStr(""),
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenTranslatorComment("translator comment"),
				tokenTranslatorComment("translator comment2"),
				tokenExtractedComment("extracted comment"),
				tokenExtractedComment("extracted comment2"),
				tokenReference("reference1"),
				tokenReference("reference2"),
				tokenReference("reference3"),
				tokenFlag("fuzzy"),
				tokenMsgCtxt("context"),
				tokenMsgctxtPreviousContext("previous context"),
				tokenMsgidPrevUntStr("msgid prev untranslated string"),
				tokenMsgID("There is 1 apple"),
				tokenMsgidPluralPrevUntStrPlural("msgid plural prev untranslated string"),
				tokenPluralMsgID("There is %d apples"),
				tokenPluralMsgStr("Il y a 1 pomme", 0),
				tokenPluralMsgStr("Il y a %d pommes", 1),
				tokenMsgID("message id"),
				tokenMsgStr("message"),
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "John Doe",
					PluralForms: pluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						TranslatorComment:     []string{"translator comment", "translator comment2"},
						ExtractedComment:      []string{"extracted comment", "extracted comment2"},
						References:            []string{"reference1", "reference2", "reference3"},
						Flag:                  "fuzzy",
						MsgCtxt:               "context",
						MsgCtxtPrevCtxt:       "previous context",
						MsgIDPrevUnt:          "msgid prev untranslated string",
						MsgID:                 "There is 1 apple",
						MsgIDPrevUntPluralStr: "msgid plural prev untranslated string",
						MsgIDPlural:           "There is %d apples",
						MsgStr:                []string{"Il y a 1 pomme", "Il y a %d pommes"},
					},
					{
						MsgID:  "message id",
						MsgStr: []string{"message"},
					},
				},
			},
		},
		{
			name: "When msgid and msgstr token values are provided",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenMsgID("message id"),
				tokenMsgStr("message"),
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "John Doe",
					PluralForms: pluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						MsgID:  "message id",
						MsgStr: []string{"message"},
					},
				},
			},
		},
		{
			name: "When plural msgid and plural msgstr token values are provided",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenMsgID("There is 1 apple"),
				tokenPluralMsgID("There is %d apples"),
				tokenPluralMsgStr("Il y a 1 pomme", 0),
				tokenPluralMsgStr("Il y a %d pommes", 1),
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "John Doe",
					PluralForms: pluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						MsgID:       "There is 1 apple",
						MsgIDPlural: "There is %d apples",
						MsgStr:      []string{"Il y a 1 pomme", "Il y a %d pommes"},
					},
				},
			},
		},
		{
			name: "Invalid plural forms format is provided",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=2"),
				tokenMsgID("There is 1 apple"),
				tokenPluralMsgID("There is %d apples"),
				tokenPluralMsgStr("Il y a 1 pomme", 0),
				tokenPluralMsgStr("Il y a %d pommes", 1),
			},
			expectedErr: fmt.Errorf("invalid plural forms format"),
		},
		{
			name: "Invalid nplurals value is provided",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=part; plural=(n != 1);"),
				tokenMsgID("There is 1 apple"),
				tokenPluralMsgID("There is %d apples"),
				tokenPluralMsgStr("Il y a 1 pomme", 0),
				tokenPluralMsgStr("Il y a %d pommes", 1),
			},
			expectedErr: fmt.Errorf("invalid nplurals value"),
		},
		{
			name: "Invalid nplurals part is provided",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("; plural=(n != 1);"),
				tokenMsgID("There is 1 apple"),
				tokenPluralMsgID("There is %d apples"),
				tokenPluralMsgStr("Il y a 1 pomme", 0),
				tokenPluralMsgStr("Il y a %d pommes", 1),
			},
			expectedErr: fmt.Errorf("invalid nplurals part"),
		},
		{
			name: "Invalid plural string order is provided",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
				tokenMsgID("There is 1 apple"),
				tokenPluralMsgID("There is %d apples"),
				tokenPluralMsgStr("Il y a 1 pomme", 1),
				tokenPluralMsgStr("Il y a %d pommes", 0),
			},
			expectedErr: fmt.Errorf("invalid plural string order %d", 1),
		},
		{
			name: "Invalid po file: no messages found",
			input: []Token{
				tokenHeaderLanguage("en-US"),
				tokenHeaderTranslator("John Doe"),
				tokenHeaderPluralForms("nplurals=2; plural=(n != 1);"),
			},
			expectedErr: fmt.Errorf("invalid po file: no messages found"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := TokensToPo(tt.input)
			if tt.expectedErr != nil {
				require.Errorf(t, err, tt.expectedErr.Error())
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}
