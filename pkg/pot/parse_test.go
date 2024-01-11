package pot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
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
				mkToken(TokenTypeMsgID, ""),
				mkToken(TokenTypeMsgStr, ""),
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
				mkToken(TokenTypeTranslatorComment, "translator comment"),
				mkToken(TokenTypeTranslatorComment, "translator comment2"),
				mkToken(TokenTypeExtractedComment, "extracted comment"),
				mkToken(TokenTypeExtractedComment, "extracted comment2"),
				mkToken(TokenTypeReference, "reference1"),
				mkToken(TokenTypeReference, "reference2"),
				mkToken(TokenTypeReference, "reference3"),
				mkToken(TokenTypeFlag, "fuzzy"),
				mkToken(TokenTypeMsgCtxt, "context"),
				mkToken(TokenTypeMsgctxtPreviousContext, "previous context"),
				mkToken(TokenTypeMsgidPrevUntStr, "msgid prev untranslated string"),
				mkToken(TokenTypeMsgID, "There is 1 apple"),
				mkToken(TokenTypeMsgidPluralPrevUntStrPlural, "msgid plural prev untranslated string"),
				mkToken(TokenTypePluralMsgID, "There is %d apples"),
				mkToken(TokenTypePluralMsgStr, "Il y a 1 pomme", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "Il y a %d pommes", withIndex(1)),
				mkToken(TokenTypeMsgID, "message id"),
				mkToken(TokenTypeMsgStr, "message"),
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "John Doe",
					PluralForms: PluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
				},
				Messages: []MessageNode{
					{
						TranslatorComment:     []string{"translator comment", "translator comment2"},
						ExtractedComment:      []string{"extracted comment", "extracted comment2"},
						References:            []string{"reference1", "reference2", "reference3"},
						Flags:                 []string{"fuzzy"},
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
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
				mkToken(TokenTypeMsgID, "message id"),
				mkToken(TokenTypeMsgStr, "message"),
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "John Doe",
					PluralForms: PluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
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
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
				mkToken(TokenTypeMsgID, "There is 1 apple"),
				mkToken(TokenTypePluralMsgID, "There is %d apples"),
				mkToken(TokenTypePluralMsgStr, "Il y a 1 pomme", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "Il y a %d pommes", withIndex(1)),
			},
			expected: Po{
				Header: HeaderNode{
					Language:    language.Make("en-US"),
					Translator:  "John Doe",
					PluralForms: PluralForm{Plural: "plural=(n != 1);", NPlurals: 2},
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
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2"),
				mkToken(TokenTypeMsgID, "There is 1 apple"),
				mkToken(TokenTypePluralMsgID, "There is %d apples"),
				mkToken(TokenTypePluralMsgStr, "Il y a 1 pomme", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "Il y a %d pommes", withIndex(1)),
			},
			expectedErr: fmt.Errorf("invalid plural forms format"),
		},
		{
			name: "Invalid nplurals value is provided",
			input: []Token{
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=part; plural=(n != 1);"),
				mkToken(TokenTypeMsgID, "There is 1 apple"),
				mkToken(TokenTypePluralMsgID, "There is %d apples"),
				mkToken(TokenTypePluralMsgStr, "Il y a 1 pomme", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "Il y a %d pommes", withIndex(1)),
			},
			expectedErr: fmt.Errorf("invalid nplurals value"),
		},
		{
			name: "Invalid nplurals part is provided",
			input: []Token{
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "; plural=(n != 1);"),
				mkToken(TokenTypeMsgID, "There is 1 apple"),
				mkToken(TokenTypePluralMsgID, "There is %d apples"),
				mkToken(TokenTypePluralMsgStr, "Il y a 1 pomme", withIndex(0)),
				mkToken(TokenTypePluralMsgStr, "Il y a %d pommes", withIndex(1)),
			},
			expectedErr: fmt.Errorf("invalid nplurals part"),
		},
		{
			name: "Invalid po file: no messages found",
			input: []Token{
				mkToken(TokenTypeHeaderLanguage, "en-US"),
				mkToken(TokenTypeHeaderTranslator, "John Doe"),
				mkToken(TokenTypeHeaderPluralForms, "nplurals=2; plural=(n != 1);"),
			},
			expectedErr: fmt.Errorf("invalid po file: no messages found"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := tokensToPo(tt.input)
			if tt.expectedErr != nil {
				require.Errorf(t, err, tt.expectedErr.Error())
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPo_Marshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
		input    Po
	}{
		{
			name: "singular",
			input: Po{
				Header: HeaderNode{
					Language:   language.English,
					Translator: "John Doe",
				},
				Messages: []MessageNode{
					{
						MsgID:            "Hello, world!",
						MsgStr:           []string{},
						Flags:            []string{"fuzzy"},
						ExtractedComment: []string{"A simple greeting"},
						References:       []string{"main.go:1"},
					},
				},
			},
			expected: `msgid ""
msgstr ""
"Language: en\n"
"Last-Translator: John Doe\n"

#: main.go:1
#. A simple greeting
#, fuzzy
msgid "Hello, world!"
msgstr ""
`,
		},
		{
			name: "plural",
			input: Po{
				Header: HeaderNode{
					Language:   language.Latvian,
					Translator: "John Doe",
					PluralForms: PluralForm{
						NPlurals: 2,
						Plural:   "n != 1",
					},
				},
				Messages: []MessageNode{
					{
						MsgID:       "There is 1 apple",
						MsgIDPlural: "There is 2 apples",
						MsgStr:      []string{"Ir 1 ābols", "Ir 2 āboli"},
					},
				},
			},
			expected: `msgid ""
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
			input: Po{
				Header: HeaderNode{
					Language:   language.Latvian,
					Translator: "John Doe",
					PluralForms: PluralForm{
						NPlurals: 2,
						Plural:   "n != 1",
					},
				},
				Messages: []MessageNode{
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
			expected: `msgid ""
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := tt.input.Marshal()

			require.Equal(t, tt.expected, string(actual))
		})
	}
}
