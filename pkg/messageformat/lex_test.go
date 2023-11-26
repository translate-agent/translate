package messageformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lex(t *testing.T) {
	t.Parallel()

	var (
		tokenEOF                 = mkToken(tokenTypeEOF, "")
		tokenExpressionOpen      = mkToken(tokenTypeExpressionOpen, "{")
		tokenExpressionClose     = mkToken(tokenTypeExpressionClose, "}")
		tokenComplexMessageOpen  = mkToken(tokenTypeComplexMessageOpen, "{{")
		tokenComplexMessageClose = mkToken(tokenTypeComplexMessageClose, "}}")
	)

	for _, test := range []struct {
		name, input string
		expected    []Token
	}{
		{
			name:  "empty quoted message",
			input: "{{}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with text with spaces",
			input: "{{ {{Hello, World!}} }}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "Hello, World!"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with text with spaces",
			input: "{{{{Hello, World!}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "Hello, World!"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable",
			input: "{{{$count}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenExpressionClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with function",
			input: "{{{:rand}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkToken(tokenTypeFunction, "rand"),
				tokenExpressionClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable and function",
			input: "{{{$guest :person}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "guest"),
				mkToken(tokenTypeFunction, "person"),
				tokenExpressionClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable, function and text",
			input: "{{{{Hello, {$guest :person} is here}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "Hello, "),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "guest"),
				mkToken(tokenTypeFunction, "person"),
				tokenExpressionClose,
				mkToken(tokenTypeText, " is here"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable, function and text",
			input: "{{{{{+button}Submit{-button}}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkToken(tokenTypeOpeningFunction, "button"),
				tokenExpressionClose,
				mkToken(tokenTypeText, "Submit"),
				tokenExpressionOpen,
				mkToken(tokenTypeClosingFunction, "button"),
				tokenExpressionClose,
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name: "plural text",
			input: "{{match {$count :number} " +
				"when 1 {{You have one notification.}} " +
				"when * {{You have {$count} notifications.}}" +
				"}}    ",
			expected: []Token{
				tokenComplexMessageOpen,
				mkToken(tokenTypeKeyword, "match"),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				tokenExpressionClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "1"),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "You have one notification."),
				tokenComplexMessageClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "*"),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "You have "),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenExpressionClose,
				mkToken(tokenTypeText, " notifications."),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
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
			name: "plural text",
			input: "{{match {$count :number} when 0 {{No notifications}}" +
				" when 1 {{You have one notification.}} " +
				"when * {{You have {$count} notifications.}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				mkToken(tokenTypeKeyword, "match"),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				tokenExpressionClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "0"),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "No notifications"),
				tokenComplexMessageClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "1"),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "You have one notification."),
				tokenComplexMessageClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "*"),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "You have "),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenExpressionClose,
				mkToken(tokenTypeText, " notifications."),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "invalid expr",
			input: "{{{$count :number",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				mkTokenErrorf("unexpected EOF"),
			},
		},
		{
			name:  "missing closing separator",
			input: "{{match {$count :number} when 1 {{You have one notification. when * {{You have {$count} notifications.}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				mkToken(tokenTypeKeyword, "match"),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				tokenExpressionClose,
				mkToken(tokenTypeKeyword, "when"),
				mkToken(tokenTypeLiteral, "1"),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "You have one notification. when * "),
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "You have "),
				tokenExpressionOpen,
				mkToken(tokenTypeVariable, "count"),
				tokenExpressionClose,
				mkToken(tokenTypeText, " notifications."),
				mkTokenErrorf("missing closing separator"),
			},
		},
		{
			name:  "invalid opening function",
			input: "{{{{{+ button}}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkTokenErrorf(`invalid first character in function name %v at %d`, 32, 7),
			},
		},
		{
			name:  "invalid closing function",
			input: "{{{{{+button}{-- button}}}}}",
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				tokenExpressionOpen,
				mkToken(tokenTypeOpeningFunction, "button"),
				tokenExpressionClose,
				tokenExpressionOpen,
				mkTokenErrorf(`invalid first character %v of function at %d`, "-", 16),
			},
		},
		{
			name:  "input with curly braces",
			input: `{{{{Chart [\{\}] was added to dashboard [\{\}]}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "Chart [{}] was added to dashboard [{}]"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "input with pipes",
			input: `{{{{Chart [\|] was added to dashboard [\|]}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "Chart [|] was added to dashboard [|]"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "input with slashes",
			input: `{{{{Chart [\\] was added to dashboard [\\]}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "Chart [\\] was added to dashboard [\\]"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "input with plus sign",
			input: `{{{{+ vēl %s}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "+ vēl %s"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "input with minus sign",
			input: `{{{{- vēl %s}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "- vēl %s"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "input with dollar sign",
			input: `{{{{$ vēl %s}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, "$ vēl %s"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
		{
			name:  "input with colon sign",
			input: `{{{{: vēl %s}}}}`,
			expected: []Token{
				tokenComplexMessageOpen,
				tokenComplexMessageOpen,
				mkToken(tokenTypeText, ": vēl %s"),
				tokenComplexMessageClose,
				tokenComplexMessageClose,
				tokenEOF,
			},
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l := lex(test.input)

			// collect tokens

			var tokens []Token

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
