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
			input: "{{+button}Submit{-button}}",
			expected: []token{
				tokenSeparatorOpen,
				tokenSeparatorOpen,
				mkToken(tokenTypeOpeningFunction, "button"),
				tokenSeparatorClose,
				mkToken(tokenTypeText, "Submit"),
				tokenSeparatorOpen,
				mkToken(tokenTypeClosingFunction, "button"),
				tokenSeparatorClose,
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "plural text",
			input: "match {$count :number} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			expected: []token{
				mkToken(tokenTypeKeyword, "match"),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				tokenSeparatorClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "1"),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "You have one notification."),
				tokenSeparatorClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "*"),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "You have "),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenSeparatorClose,
				mkToken(tokenTypeText, " notifications."),
				tokenSeparatorClose,
				tokenEOF,
			},
		},

		/*
			Tree{
				RootNode: Node{
					keywords:
			}
		*/
		{
			name:  "plural text",
			input: "match {$count :number} when 0 {No notifications} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			expected: []token{
				mkToken(tokenTypeKeyword, "match"),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				tokenSeparatorClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "0"),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "No notifications"),
				tokenSeparatorClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "1"),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "You have one notification."),
				tokenSeparatorClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "*"),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "You have "),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenSeparatorClose,
				mkToken(tokenTypeText, " notifications."),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "invalid expr",
			input: "{$count :number",
			expected: []token{
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				mkTokenErrorf("unexpected EOF"),
			},
		},
		{
			name:  "missing closing separator",
			input: "match {$count :number} when 1 {You have one notification. when * {You have {$count} notifications.}",
			expected: []token{
				mkToken(tokenTypeKeyword, "match"),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				tokenSeparatorClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "1"),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "You have one notification. when * "),
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "You have "),
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenSeparatorClose,
				mkToken(tokenTypeText, " notifications."),
				tokenSeparatorClose,
				mkTokenErrorf("missing closing separator"),
			},
		},
		{
			name:  "invalid variable",
			input: "{$ count :number}",
			expected: []token{
				tokenSeparatorOpen,
				mkTokenErrorf(`invalid first character %s in variable at %d`, " ", 3),
			},
		},
		{
			name:  "invalid function",
			input: "{$count : number}",
			expected: []token{
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				mkTokenErrorf(`invalid first character %s in function at %d`, " ", 10),
			},
		},
		{
			name:  "invalid opening function",
			input: "{{+ button}}",
			expected: []token{
				tokenSeparatorOpen,
				tokenSeparatorOpen,
				mkTokenErrorf(`invalid first character in function name %v at %d`, 32, 4),
			},
		},
		{
			name:  "invalid closing function",
			input: "{{+button}{-- button}}",
			expected: []token{
				tokenSeparatorOpen,
				tokenSeparatorOpen,
				mkToken(tokenTypeOpeningFunction, "button"),
				tokenSeparatorClose,
				tokenSeparatorOpen,
				mkTokenErrorf(`invalid first character %v of function at %d`, "-", 13),
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
