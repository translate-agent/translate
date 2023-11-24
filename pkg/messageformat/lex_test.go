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
		name     string
		input    string // MessageFormat2 formatted string
		expected []Token
	}{
		{
			name:     "empty",
			input:    "",
			expected: []Token{tokenEOF},
		},
		{
			name:     "empty expr",
			input:    "{}",
			expected: []Token{tokenSeparatorOpen, tokenSeparatorClose, tokenEOF},
		},
		{
			name:     "expr with text",
			input:    "{Hello, World!}",
			expected: []Token{tokenSeparatorOpen, mkToken(tokenTypeText, "Hello, World!"), tokenSeparatorClose, tokenEOF},
		},
		{
			name:     "expr with variable",
			input:    "{$count}",
			expected: []Token{tokenSeparatorOpen, mkToken(tokenTypeVariable, "count"), tokenSeparatorClose, tokenEOF},
		},
		{
			name:     "expr with function",
			input:    "{:rand}",
			expected: []Token{tokenSeparatorOpen, mkToken(tokenTypeFunction, "rand"), tokenSeparatorClose, tokenEOF},
		},
		{
			name:  "expr with variable and function",
			input: "{$guest :person}",
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "guest"),
				mkToken(tokenTypeFunction, "person"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "expr with variable, function and text",
			input: "{Hello, {$guest :person} is here}",
			expected: []Token{
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
			name:  "expr with variable, function and text",
			input: "{{+button}Submit{-button}}",
			expected: []Token{
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
			expected: []Token{
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
			name: "plural text",
			input: "match {$count :number} when 0 {No notifications}" +
				" when 1 {You have one notification.} " +
				"when * {You have {$count} notifications.}",
			expected: []Token{
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
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeVariable, "count"),
				mkToken(tokenTypeFunction, "number"),
				mkTokenErrorf("unexpected EOF"),
			},
		},
		{
			name:  "missing closing separator",
			input: "match {$count :number} when 1 {You have one notification. when * {You have {$count} notifications.}",
			expected: []Token{
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
			name:  "invalid opening function",
			input: "{{+ button}}",
			expected: []Token{
				tokenSeparatorOpen,
				tokenSeparatorOpen,
				mkTokenErrorf(`invalid first character in function name %v at %d`, 32, 4),
			},
		},
		{
			name:  "invalid closing function",
			input: "{{+button}{-- button}}",
			expected: []Token{
				tokenSeparatorOpen,
				tokenSeparatorOpen,
				mkToken(tokenTypeOpeningFunction, "button"),
				tokenSeparatorClose,
				tokenSeparatorOpen,
				mkTokenErrorf(`invalid first character %v of function at %d`, "-", 13),
			},
		},
		{
			name:  "input with curly braces",
			input: `{Chart [\{\}] was added to dashboard [\{\}]}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "Chart [\\{\\}] was added to dashboard [\\{\\}]"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "input with pipes",
			input: `{Chart [\|] was added to dashboard [\|]}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "Chart [\\|] was added to dashboard [\\|]"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "input with slashes",
			input: `{Chart [\\] was added to dashboard [\\]}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "Chart [\\\\] was added to dashboard [\\\\]"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "input with plus sign",
			input: `{+ vēl \%s}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "+ vēl \\%s"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "input with minus sign",
			input: `{- vēl \%s}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "- vēl \\%s"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "input with dollar sign",
			input: `{$ vēl \%s}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "$ vēl \\%s"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name:  "input with colon sign",
			input: `{: vēl \%s}`,
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, ": vēl \\%s"),
				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name: "python old format placeholder",
			// %(object)s does not exist in this database.
			input: "{{:Placeholder name=object format=pythonVar type=string} does not exist in this database.}",
			expected: []Token{
				tokenSeparatorOpen,

				tokenSeparatorOpen,
				mkToken(tokenTypeFunction, "Placeholder"),
				mkToken(tokenTypeOption, "name=object"),
				mkToken(tokenTypeOption, "format=pythonVar"),
				mkToken(tokenTypeOption, "type=string"),
				tokenSeparatorClose,

				mkToken(tokenTypeText, " does not exist in this database."),

				tokenSeparatorClose,
				tokenEOF,
			},
		},
		{
			name: "empty brackets placeholder with escape chars",
			// <{}|Explore in Superset>\n
			input: "{\\<{:Placeholder format=emptyBracket}\\|Explore in Superset\\>\n}",
			expected: []Token{
				tokenSeparatorOpen,
				mkToken(tokenTypeText, "\\<"),

				tokenSeparatorOpen,
				mkToken(tokenTypeFunction, "Placeholder"),
				mkToken(tokenTypeOption, "format=emptyBracket"),
				tokenSeparatorClose,

				mkToken(tokenTypeText, "\\|Explore in Superset\\>\n"),

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
