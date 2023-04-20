package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestLexMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedErr error
		expected    []Token
	}{
		{
			name:  "simple message",
			input: "{Hello, world!}",
			expected: []Token{
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "Hello, world!"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeEOF},
			},
		},
		{
			name:  "simple message with escapes",
			input: "{Hello, world!} \n\t ",
			expected: []Token{
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "Hello, world!"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeEOF},
			},
		},

		{
			name:  "message with variable",
			input: "{Hello, {$userName}!}",
			expected: []Token{
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "Hello, "},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "userName"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeText, Value: "!"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeEOF},
			},
		},
		{
			name:  "message with plurals",
			input: "match {$count :number} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			expected: []Token{
				{Type: TokenTypeKeyword, Value: "match"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "count"},
				{Type: TokenTypeFunction, Value: "number"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "1"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "You have one notification."},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "*"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "You have "},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "count"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeText, Value: " notifications."},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeEOF},
			},
		},
		{
			name:  "message with UTF characters",
			input: "match {$日本語 :number} when 1 {日本語 日本語} when * {日本語 {$日本語} 日本語.}",
			expected: []Token{
				{Type: TokenTypeKeyword, Value: "match"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "日本語"},
				{Type: TokenTypeFunction, Value: "number"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "1"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "日本語 日本語"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "*"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "日本語 "},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "日本語"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeText, Value: " 日本語."},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeEOF},
			},
		},
		{
			name: "message with multiple plurals",
			input: "match {$count :number} when 0 {You have no notifications.} " +
				"when 1 {You have one notification.} " +
				"when * {You have {$count} notifications.}",
			expected: []Token{
				{Type: TokenTypeKeyword, Value: "match"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "count"},
				{Type: TokenTypeFunction, Value: "number"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "0"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "You have no notifications."},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "1"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "You have one notification."},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeKeyword, Value: "when"},
				{Type: TokenTypeLiteral, Value: "*"},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeText, Value: "You have "},
				{Type: TokenTypeSeparatorOpen, Value: "{"},
				{Type: TokenTypeVariable, Value: "count"},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeText, Value: " notifications."},
				{Type: TokenTypeSeparatorClose, Value: "}"},
				{Type: TokenTypeEOF},
			},
		},
		{
			name:        "variable starts with space",
			input:       "{Hello, {$ userName}!}",
			expectedErr: fmt.Errorf(`variable does not starts with "$"`),
		},
		{
			name:        "message with plurals with invalid function",
			input:       "match {$count : number} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			expectedErr: fmt.Errorf(`function does not starts with ":"`),
		},
		{
			name:        "message with UTF characters",
			input:       "{Hello, {$ 日本語}!}",
			expectedErr: fmt.Errorf(`variable does not starts with "$"`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := Lex(tt.input)

			if tt.expectedErr != nil {
				assert.Errorf(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}
