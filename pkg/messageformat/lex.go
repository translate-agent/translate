package messageformat

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TODO(jhorsts): use cursor position by line number and position in a line

const eof = -1

type tokenType int

const (
	tokenTypeError tokenType = iota
	tokenTypeEOF
	tokenTypeVariable
	tokenTypeFunction
	tokenTypeExpressionOpen
	tokenTypeExpressionClose
	tokenTypeComplexMessageOpen
	tokenTypeComplexMessageClose
	tokenTypeText
	tokenTypeOpeningFunction
	tokenTypeClosingFunction
	tokenTypeKeyword
	tokenTypeLiteral
	tokenTypeOption
)

const (
	KeywordMatch = "match"
	KeywordWhen  = "when"
)

type Token struct {
	val string
	typ tokenType
}

func mkToken(typ tokenType, val string) Token {
	return Token{typ: typ, val: val}
}

func mkTokenErrorf(s string, args ...interface{}) Token {
	return Token{typ: tokenTypeError, val: fmt.Sprintf(s, args...)}
}

func lex(input string) *lexer {
	return &lexer{
		// remove spaces outside messageFormat v2
		input: strings.TrimSpace(input),
	}
}

type lexer struct {
	input        string
	prevToken    *Token
	token        Token
	pos          int
	exprDepth    int
	textToFollow bool
	whenFound    bool
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

func (l *lexer) nextToken() Token {
	l.emitToken(mkToken(tokenTypeEOF, ""))

	state := lexOutsideExpr

	if l.exprDepth > 0 {
		state = lexExpr
	}

	for {
		state := state(l)

		if state == nil {
			l.prevToken = &Token{typ: l.token.typ, val: l.token.val}
			return l.token
		}
	}
}

func (l *lexer) emitToken(t Token) stateFn {
	l.token = t

	return nil
}

type stateFn func(*lexer) stateFn

func lexOutsideExpr(l *lexer) stateFn {
	minLength := 4

	// optional check for the input length
	if len(l.input) < minLength {
		return l.emitToken(mkToken(tokenTypeError, "input can't be shorter then 4 chars '{{}}'"))
	}

	var expr string

	// check start of the input is {{ and shift pos to skip these chars,
	// else create error token
	// usage: default case to check start of input
	if l.pos == 0 {
		if strings.HasPrefix(l.input, "{{") {
			l.next() // skip "{" rune
			l.next() // skip "{" rune

			return l.emitToken(mkToken(tokenTypeComplexMessageOpen, "{{"))
		}

		return l.emitToken(mkToken(tokenTypeError, "syntax error on opening complex message"))
	}

	for {
		v := l.next()

		if v == eof {
			return l.emitToken(mkToken(tokenTypeEOF, ""))
		}

		if v == '{' {
			l.exprDepth++

			// if current and next "v" equals to {{ shift pos
			// usage: wrap TokenTypeText
			if l.peek() == '{' {
				l.next()

				return l.emitToken(mkToken(tokenTypeComplexMessageOpen, "{{"))
			}

			return l.emitToken(mkToken(tokenTypeExpressionOpen, "{"))
		}

		expr += string(v)

		if strings.TrimSpace(expr) == KeywordMatch {
			return l.emitToken(mkToken(tokenTypeKeyword, KeywordMatch))
		}

		if strings.TrimSpace(expr) == KeywordWhen {
			l.whenFound = true
			return l.emitToken(mkToken(tokenTypeKeyword, KeywordWhen))
		}

		if l.whenFound {
			l.whenFound = false
			return lexLiteral(l)
		}

		// check for closing of complex message outside expression
		if v == '}' && l.peek() == '}' {
			l.next()

			if l.peek() == eof {
				return l.emitToken(mkToken(tokenTypeComplexMessageClose, "}}"))
			}
		}
	}
}

func lexLiteral(l *lexer) stateFn {
	var literal string

	for {
		v := l.next()

		if l.peek() == '{' {
			return l.emitToken(mkToken(tokenTypeLiteral, strings.TrimSpace(literal)))
		}

		literal += string(v)
	}
}

func lexText(l *lexer) stateFn {
	l.textToFollow = false

	var sb strings.Builder

	for {
		v := l.next()

		// EOF is not supposed to be in the text.
		if v == eof {
			return l.emitToken(mkTokenErrorf("unexpected EOF"))
		}

		// Write the character to the buffer.
		sb.WriteRune(v)

		// If we just wrote a backslash, write the next character too.
		if v == '\\' {
			sb.WriteRune(l.next())
		}

		if l.peek() == '}' || l.peek() == '{' {
			return l.emitToken(mkToken(tokenTypeText, sb.String()))
		}
	}
}

func lexExpr(l *lexer) stateFn {
	v := l.next()

	if v == eof {
		if l.exprDepth > 0 {
			return l.emitToken(mkTokenErrorf("missing closing separator"))
		}

		return nil
	}

	switch v {
	default:
		if (l.prevToken.typ == tokenTypeFunction || l.prevToken.typ == tokenTypeOption) && v == ' ' {
			return lexOption(l)
		}

		if isSpace(v) && !l.textToFollow {
			l.nextToken()
			return nil
		} else {
			l.backup()

			return lexText(l)
		}
	case '$':
		return processChar(l, lexVariable)
	case ':':
		return processChar(l, lexFunction)
	case '+':
		return processFunction(l, lexOpeningFunction)
	case '-':
		return processFunction(l, lexClosingFunction)
	case '{':
		l.exprDepth++

		// set "{{" before text token
		if isAlpha(l.peek()) || isSpace(l.peek()) || l.peek() == '{' {
			l.next() // skip "{" rune
			l.emitToken(mkToken(tokenTypeComplexMessageOpen, "{{"))
		} else {
			l.emitToken(mkToken(tokenTypeExpressionOpen, "{"))
		}

		return nil
	case '}':
		l.exprDepth--
		l.textToFollow = true

		// close text token with "}}" or exp with "}"
		if l.prevToken.typ == tokenTypeText || l.prevToken.typ == tokenTypeExpressionClose ||
			(l.prevToken.typ == tokenTypeComplexMessageClose && l.exprDepth == 0) {
			if l.exprDepth > 0 {
				l.emitToken(mkToken(tokenTypeError, "missing closing separator"))
			} else {
				l.next() // skip "}" rune
				l.emitToken(mkToken(tokenTypeComplexMessageClose, "}}"))
			}
		} else {
			l.emitToken(mkToken(tokenTypeExpressionClose, "}"))
		}

		return nil
	}
}

func lexOpeningFunction(l *lexer) stateFn {
	first := l.next()

	if !isNameFirstChar(first) {
		return l.emitToken(mkTokenErrorf(`invalid first character in function name %v at %d`, first, l.pos))
	}

	function := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeOpeningFunction, function))
		}

		function += string(v)
	}
}

func lexClosingFunction(l *lexer) stateFn {
	first := l.next()

	if !isNameFirstChar(first) {
		return l.emitToken(mkTokenErrorf(`invalid first character %s of function at %d`, string(first), l.pos))
	}

	function := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeClosingFunction, function))
		}

		function += string(v)
	}
}

func lexFunction(l *lexer) stateFn {
	first := l.next()

	function := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeFunction, function))
		}

		function += string(v)
	}
}

func lexOption(l *lexer) stateFn {
	var sb strings.Builder

	for {
		v := l.next()

		if v == ' ' || v == '}' {
			l.backup()
			return l.emitToken(mkToken(tokenTypeOption, sb.String()))
		}

		sb.WriteRune(v)
	}
}

func lexVariable(l *lexer) stateFn {
	first := l.next()

	variable := string(first)

	for {
		v := l.next()

		if !isNameChar(v) {
			l.backup()
			return l.emitToken(mkToken(tokenTypeVariable, variable))
		}

		variable += string(v)
	}
}

func processFunction(l *lexer, f func(*lexer) stateFn) stateFn {
	if l.exprDepth > 1 {
		return f(l)
	}

	l.backup()

	return lexText(l)
}

func processChar(l *lexer, f func(*lexer) stateFn) stateFn {
	if isAlpha(l.peek()) {
		return f(l)
	}

	l.backup()

	return lexText(l)
}

// helpers

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
