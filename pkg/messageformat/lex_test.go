package messageformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lex(t *testing.T) {
	t.Parallel()

	var (
		tokenEOF            = mkToken(tokenTypeEOF, "")
		tokenSeparatorOpen  = mkToken(tokenTypeSeparatorOpen, "{")
		tokenSeparatorClose = mkToken(tokenTypeSeparatorClose, "}")
	)

	for _, test := range []struct {
		name, input string
		expected    []token
	}{
		{
			name:     "empty",
			input:    "",
			expected: []token{tokenEOF},
		},
		{
			name:     "empty expr",
			input:    "{}",
			expected: []token{tokenSeparatorOpen, tokenSeparatorClose, tokenEOF},
		},
		{
			name:     "expr with text",
			input:    "{Hello, World!}",
			expected: []token{tokenSeparatorOpen, mkToken(tokenTypeText, "Hello, World!"), tokenSeparatorClose, tokenEOF},
		},
		{
			name:     "expr with variable",
			input:    "{$count}",
			expected: []token{tokenSeparatorOpen, mkToken(tokenTypeVariable, "count"), tokenSeparatorClose, tokenEOF},
		},
		{
			name:     "expr with function",
			input:    "{:rand}",
			expected: []token{tokenSeparatorOpen, mkToken(tokenTypeFunction, "rand"), tokenSeparatorClose, tokenEOF},
		},
		{
			name:  "expr with variable and function",
			input: "{$guest :person}",
			expected: []token{
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "guest"),
				mkToken(tokenTypeFunction, "person"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable and function and text",
			input: "{Hello, {$guest :person} is here}",
			expected: []token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "Hello, "),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "guest"),
				mkToken(tokenTypeFunction, "person"),
				tokenSeparatorClose,
				mkToken(tokenTypeText, " is here"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable and function and text",
			input: "{+button}Submit{-button}",
			expected: []token{
				tokenSeparatorOpen,
				mkToken(tokenTypeOpeningFunction, "button"),
				tokenSeparatorClose,
				mkToken(tokenTypeText, "Submit"),
				tokenSeparatorOpen,
				mkToken(tokenTypeClosingFunction, "button"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l := lex(test.input)

			// collect tokens

			var tokens []token

			for {
				v := l.nextToken()

				tokens = append(tokens, v)

				if v.typ == tokenTypeEOF || v.typ == tokenTypeError {
					break
				}
			}

			// assert

			assert.Equal(t, test.expected, tokens)
		})
	}
}
