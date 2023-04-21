package message

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Token struct {
	Value string
	Type  TokenType
}

type TokenType int

const (
	TokenTypeUnknown TokenType = iota
	TokenTypeKeyword
	TokenTypeSeparatorOpen
	TokenTypeSeparatorClose
	TokenTypeLiteral
	TokenTypeText
	TokenTypeFunction
	TokenTypeVariable
	TokenTypeEOF
)

const (
	KeywordMatch   = "match"
	KeywordWhen    = "when"
	Dollar         = "$"
	Colon          = ":"
	SeparatorOpen  = '{'
	SeparatorClose = '}'
	EOF            = rune(0)
)

type lexer struct {
	str []rune
	pos int
}

func Lex(str string) ([]Token, error) {
	l := lexer{str: []rune(str)}

	return l.parse()
}

func (l *lexer) parse() ([]Token, error) {
	var tokens []Token

	var nextTokenType string

	var textToFollow bool

	for !l.isEOF() {
		v := l.current()

		s := string(l.str[l.pos:])

		switch {
		default:
			textToFollow = false

			text, err := l.parseText()
			if err != nil {
				return nil, fmt.Errorf("parse text: %w", err)
			}

			if l.lookup(l.pos) != SeparatorClose {
				textToFollow = true
			}

			tokens = append(tokens, text)
		case l.isWhitespace(v) && !textToFollow:
			if l.isEOF() {
				return tokens, nil
			}

			l.pos++
			// noop
		case nextTokenType == "literal":
			tokens = append(tokens, l.parseLiteral())
			nextTokenType = ""
		case strings.HasPrefix(s, Dollar):
			variable, err := l.parseVariable()
			if err != nil {
				return nil, fmt.Errorf("parse variable: %w", err)
			}

			tokens = append(tokens, variable)
		case strings.HasPrefix(s, Colon):
			function, err := l.parseFunction()
			if err != nil {
				return nil, fmt.Errorf("parse function: %w", err)
			}

			tokens = append(tokens, function)
		case v == SeparatorOpen:
			tokens = append(tokens, Token{Type: TokenTypeSeparatorOpen, Value: "{"})
			l.pos++
		case v == SeparatorClose:
			tokens = append(tokens, Token{Type: TokenTypeSeparatorClose, Value: "}"})
			l.pos++
		case strings.HasPrefix(s, KeywordMatch):
			tokens = append(tokens, Token{Type: TokenTypeKeyword, Value: KeywordMatch})

			l.pos += len(KeywordMatch)
		case strings.HasPrefix(s, KeywordWhen):
			tokens = append(tokens, Token{Type: TokenTypeKeyword, Value: KeywordWhen})

			l.pos += len(KeywordWhen)

			nextTokenType = "literal"
		}
	}

	return append(tokens, Token{Type: TokenTypeEOF}), nil
}

func (l *lexer) parseFunction() (Token, error) {
	function := Token{Type: TokenTypeFunction}

	l.pos++

	if l.current() == ' ' {
		return Token{}, errors.New(`function does not start with ":"`)
	}

	for {
		v := l.current()

		if v == SeparatorClose {
			break
		}

		function.Value += string(v)

		l.pos++
	}

	return function, nil
}

func (l *lexer) parseText() (Token, error) {
	text := Token{Type: TokenTypeText}

	for {
		if l.isEOF() {
			return Token{}, errors.New(`text does not end with "}"`)
		}

		v := l.current()

		if v == SeparatorOpen || v == SeparatorClose {
			return text, nil
		}

		text.Value += string(v)

		l.pos++
	}
}

func (l *lexer) parseLiteral() Token {
	literal := Token{Type: TokenTypeLiteral}

	for {
		v := l.current()

		if v == SeparatorOpen {
			break
		}

		literal.Value += strings.TrimSpace(string(v))

		l.pos++
	}

	return literal
}

// parseVariable parses variable name according to
// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#names.
// name    = name-start *name-char ; matches XML https://www.w3.org/TR/xml/#NT-Name
// name-start = ALPHA / "_"
//
//	/ %xC0-D6 / %xD8-F6 / %xF8-2FF
//	/ %x370-37D / %x37F-1FFF / %x200C-200D
//	/ %x2070-218F / %x2C00-2FEF / %x3001-D7FF
//	/ %xF900-FDCF / %xFDF0-FFFD / %x10000-EFFFF
//
// name-char = name-start / DIGIT / "-" / "." / %xB7
//
//	/ %x0300-036F / %x203F-2040
func (l *lexer) parseVariable() (Token, error) {
	variable := Token{Type: TokenTypeVariable}

	l.pos++

	if l.current() == ' ' {
		return Token{}, errors.New(`variable does not start with "$"`)
	}

	for {
		v := l.current()

		if !unicode.IsLetter(v) && v != '_' || l.isEOF() {
			return variable, nil
		}

		variable.Value += string(v)

		l.pos++
	}
}

func (l *lexer) current() rune {
	return l.str[l.pos]
}

func (l *lexer) lookup(i int) rune {
	return l.str[i]
}

func (l *lexer) isEOF() bool {
	return len(l.str) <= l.pos
}

func (l *lexer) isWhitespace(v rune) bool {
	return v == ' ' || v == '\t' || v == '\n'
}
