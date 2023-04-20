package message

import (
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
	KeywordLet     = "let"
	KeywordWhen    = "when"
	Dollar         = "$"
	Colon          = ":"
	Plus           = '+'
	Minus          = '-'
	SeparatorOpen  = '{'
	SeparatorClose = '}'
	EOF            = rune(0)
)

type lexer struct {
	str []rune
	pos int
}

func (l *lexer) current() rune {
	return l.str[l.pos]
}

func (l *lexer) next() rune {
	l.pos++

	if l.pos == len(l.str) {
		return EOF
	}

	return l.str[l.pos]
}

func (l *lexer) nextNotWhitespace() rune {
	for {
		v := l.current()

		if l.isEOF() {
			return EOF
		}

		if !l.isWhitespace(v) {
			return v
		}

		l.pos++
	}
}

func (l *lexer) lookup(i int) rune {
	return l.str[i]
}

func (l *lexer) isEOF() bool {
	return len(l.str) <= l.pos
}

// isAlpha returns true if v is alphabetic character.
func (l *lexer) isAlpha(v rune) bool {
	return ('a' <= v && v <= 'z') || ('A' <= v && v <= 'Z')
}

func (l *lexer) isWhitespace(v rune) bool {
	return v == ' ' || v == '\t' || v == '\n'
}

func (l *lexer) parse() ([]Token, error) {
	var tokens []Token

	var nextTokenType string

	for !l.isEOF() {
		v := l.current()

		s := string(l.str[l.pos:])

		switch {
		default:
			text, err := l.parseText()
			if err != nil {
				return nil, err
			}

			tokens = append(tokens, text)
		case l.isWhitespace(v):
			if l.isEOF() {
				return tokens, nil
			}
			l.pos++
			// noop
		case nextTokenType == "literal":
			literal, err := l.parseLiteral()
			if err != nil {
				return nil, err
			}

			tokens = append(tokens, literal)
			nextTokenType = ""
		case strings.HasPrefix(s, Dollar):
			variable, err := l.parseVariable()
			if err != nil {
				return nil, err
			}

			tokens = append(tokens, variable)
		case strings.HasPrefix(s, Colon):
			function, err := l.parseFunction()
			if err != nil {
				return nil, err
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
		v := l.current()

		if v == SeparatorOpen || v == SeparatorClose {
			return text, nil
		}

		text.Value += string(v)

		l.pos++
	}
}

func (l *lexer) parseLiteral() (Token, error) {
	literal := Token{Type: TokenTypeLiteral}

	for {
		v := l.current()

		if v == SeparatorOpen {
			break
		}

		literal.Value += strings.TrimSpace(string(v))

		l.pos++
	}

	return literal, nil
}

// parseVariable parses variable name according to https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#names.
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

	for {
		v := l.current()

		if !unicode.IsLetter(v) && v != '_' || l.isEOF() {
			return variable, nil
		}

		variable.Value += string(v)

		l.pos++
	}
}

func Lex(str string) ([]Token, error) {
	l := lexer{str: []rune(str)}

	return l.parse()
}

// combineTextTokens combining Text tokens into one sentence.
// func combineTextTokens(tokens, parsedTokens []Token) ([]Token, error) {
//	var txt strings.Builder
//
//	for i := 0; i < len(tokens); i++ {
//		if tokens[i].Type == TokenTypeText {
//			if _, err := txt.WriteString(tokens[i].Value); err != nil {
//				return []Token{}, errors.New("write Text token")
//			}
//
//			if i+1 < len(tokens) && tokens[i+1].Type != TokenTypeText {
//				parsedTokens = append(parsedTokens, Token{Type: TokenTypeText, Value: txt.String()})
//
//				txt.Reset()
//			}
//		} else {
//			parsedTokens = append(parsedTokens, tokens[i])
//		}
//	}
//
//	return parsedTokens, nil
// }

// createTokensFromBuffer breaks the input text into tokens that can be processed separately.
// func createTokensFromBuffer(buffer []rune, placeholderLevel int) []Token {
//	var newTokens []Token
//
//	re := regexp.MustCompile(`\\[^\S\n]?`)
//	v := re.ReplaceAllString(strings.TrimSpace(string(buffer)), "")
//
//	if v != "" {
//		switch v {
//		case KeywordMatch, KeywordLet, KeywordWhen:
//			newTokens = append(newTokens, Token{Type: TokenTypeKeyword, Value: v})
//		default:
//			if placeholderLevel == 0 {
//				newTokens = append(newTokens, Token{Type: TokenTypeLiteral, Value: v})
//			} else {
//				switch buffer[0] {
//				case Dollar, Plus, Minus:
//					if placeholderLevel > 0 {
//						newTokens = append(newTokens,
//							Token{
//								Type:  TokenTypeVariable,
//								Value: strings.ReplaceAll(v, "$", ""),
//							})
//					}
//				case Colon:
//					if placeholderLevel > 0 {
//						newTokens = append(newTokens, Token{
//							Type:  TokenTypeFunction,
//							Value: strings.ReplaceAll(v, ":", ""),
//						})
//					}
//				default:
//					newTokens = append(newTokens, Token{Type: TokenTypeText, Value: string(buffer)})
//				}
//			}
//		}
//	}
//
//	return newTokens
// }
