package messageformat

import (
	"unicode/utf8"
)

const eof = -1

type tokenType int

var textToFollow bool

const (
	tokenTypeError tokenType = iota
	tokenTypeEOF
	tokenTypeVariable
	tokenTypeFunction
	tokenTypeSeparatorOpen
	tokenTypeSeparatorClose
	tokenTypeText
	tokenTypeOpeningFunction
	tokenTypeClosingFunction
)

type token struct {
	val string
	typ tokenType
}

func mkToken(typ tokenType, val string) token {
	return token{typ: typ, val: val}
}

func lex(input string) *lexer {
	return &lexer{input: input}
}

type lexer struct {
	input string
	token token
	pos   int
}

// next returns the next rune.
func (l *lexer) next() rune {
	if len(l.input) <= l.pos {
		return eof
	}

	r, n := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += n

	return r
}

func (l *lexer) backup() {
	_, n := utf8.DecodeLastRuneInString(l.input[:l.pos])
	l.pos -= n
}

// peek peeks at the next rune.
func (l *lexer) peek() rune {
	pos := l.pos
	r := l.next()
	l.pos = pos

	return r
}

func (l *lexer) nextToken() token {
	l.token = mkToken(tokenTypeEOF, "")

	state := lexExpr

	for {
		state := state(l)

		if state == nil {
			return l.token
		}
	}
}

func (l *lexer) emitToken(t token) stateFn {
	l.token = t

	return nil
}

type stateFn func(*lexer) stateFn

func lexText(l *lexer) stateFn {
	textToFollow = false

	var s string

	for {
		v := l.next()

		if v == eof {
			return l.emitToken(mkToken(tokenTypeError, ""))
		}

		s += string(v)

		if l.peek() == '}' || l.peek() == '{' {
			return l.emitToken(mkToken(tokenTypeText, s))
		}
	}
}

func lexExpr(l *lexer) stateFn {
	v := l.next()

	if v == eof {
		return nil
	}

	switch v {
	default:
		if isSpace(v) && !textToFollow {
			l.nextToken()
			return nil
		} else {
			l.backup()

			return lexText(l)
		}
	case '$':
		return lexVariable(l)
	case ':':
		return lexFunction(l)
	case '+':
		return lexOpeningFunction(l)
	case '-':
		return lexClosingFunction(l)
	case '{':
		l.token = mkToken(tokenTypeSeparatorOpen, "{")

		return nil
	case '}':
		textToFollow = true

		l.token = mkToken(tokenTypeSeparatorClose, "}")

		return nil
	}
}

func lexMatch(l *lexer) stateFn {
	return nil
}
func lexOpeningFunction(l *lexer) stateFn {
	first := l.next()

	if !isNameFirstChar(first) {
		return l.emitToken(mkToken(tokenTypeError, ""))
	}

	s := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeOpeningFunction, s))
		}

		s += string(v)
	}
}

func lexClosingFunction(l *lexer) stateFn {
	first := l.next()

	if !isNameFirstChar(first) {
		return l.emitToken(mkToken(tokenTypeError, ""))
	}

	s := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeClosingFunction, s))
		}

		s += string(v)
	}
}

func lexFunction(l *lexer) stateFn {
	first := l.next()

	if !isNameFirstChar(first) {
		return l.emitToken(mkToken(tokenTypeError, ""))
	}

	s := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeFunction, s))
		}

		s += string(v)
	}
}

func lexVariable(l *lexer) stateFn {
	first := l.next()

	if !isNameFirstChar(first) {
		return l.emitToken(mkToken(tokenTypeError, ""))
	}

	s := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeVariable, s))
		}

		s += string(v)
	}
}

// isAlpha returns true if v is alphabetic character.
func isAlpha(v rune) bool {
	return ('a' <= v && v <= 'z') || ('A' <= v && v <= 'Z')
}

// isNameFirstChar returns true if the first character v is allowed character according
// to https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#names
//
// name-start = ALPHA / "_"
//
//	/ %xC0-D6 / %xD8-F6 / %xF8-2FF
//	/ %x370-37D / %x37F-1FFF / %x200C-200D
//	/ %x2070-218F / %x2C00-2FEF / %x3001-D7FF
//	/ %xF900-FDCF / %xFDF0-FFFD / %x10000-EFFFF
//
//nolint:gocognit
func isNameFirstChar(v rune) bool {
	return isAlpha(v) ||
		v == '_' ||
		0xC0 <= v && v <= 0xD6 ||
		0xD8 <= v && v <= 0xF6 ||
		0xF8 <= v && v <= 0x2FF ||
		0x370 <= v && v <= 0x37D ||
		0x37F <= v && v <= 0x1FFF ||
		0x2C00 <= v && v <= 0x2FEF ||
		0x3001 <= v && v <= 0xD7FF ||
		0xF900 <= v && v <= 0xFDCF ||
		0xFDF0 <= v && v <= 0xFFFD ||
		0x10000 <= v && v <= 0xEFFFF
}

// isNameChar returns true if v is allowed character according
// to https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#names
//
// name-char = name-start / DIGIT / "-" / "." / %xB7 / %x0300-036F / %x203F-2040.
func isNameChar(v rune) bool {
	return isAlpha(v) ||
		'0' <= v && v <= '9' ||
		v == '-' ||
		v == '.' ||
		v == 0xB7 ||
		0x0300 <= v && v <= 0x036F ||
		0x203F <= v && v <= 2040
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}
