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
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "Hello, world!", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
		},
		{
			name:  "message with variable",
			input: "{Hello, {$userName}!}",
			expected: []Token{
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "Hello, ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "userName", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: "!", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
		},
		{
			name:  "message with plurals",
			input: "match {$count :number} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			expected: []Token{
				{Type: Keyword, Value: "match ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "count ", Level: 1},
				{Type: Function, Value: "number", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "1 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have one notification.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "* ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "count", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: " notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
		},
		{
			name:  "message with non-ASCII characters ",
			input: "match {$日 :本} when 1 {日本語 日本語} when * {日本語 {$日} 日本語.}",
			expected: []Token{
				{Type: Keyword, Value: "match ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "日 ", Level: 1},
				{Type: Function, Value: "本", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "1 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "日本語 日本語", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "* ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "日本語 ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "日", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: " 日本語.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
		},
		{
			name: "message with multiple plurals",
			input: "match {$count :number} when 0 {You have no notifications.} " +
				"when 1 {You have one notification.} " +
				"when * {You have {$count} notifications.}",
			expected: []Token{
				{Type: Keyword, Value: "match ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "count ", Level: 1},
				{Type: Function, Value: "number", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "0 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have no notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "1 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have one notification.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "* ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "count", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: " notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
		},
		{
			name:        "message with Variable",
			input:       "{Hello, {$ userName}!}",
			expectedErr: fmt.Errorf("variable should not contains space"),
		},
		{
			name:        "message with plurals",
			input:       "match {$count : number} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			expectedErr: fmt.Errorf("variable should not contains space"),
		},
		{
			name:        "message with non-ASCII characters",
			input:       "{Hello, {$ 日本語}!}",
			expectedErr: fmt.Errorf("variable should not contains space"),
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
