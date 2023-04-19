package message

import (
	"errors"
	"regexp"
	"strings"
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
	Dollar         = '$'
	Colon          = ':'
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

func (l *lexer) lookup(i int) rune {
	return l.str[i]
}

func (l *lexer) isEOF() bool {
	return len(l.str) <= l.pos
}

func (l *lexer) parse() ([]Token, error) {
	var tokens []Token

	for !l.isEOF() {
		v := l.current()

		switch v {
		default:
			l.pos++
		case SeparatorOpen:
			l.pos++

			if l.current() == Dollar {
				tokens = append(tokens, l.parseVariable()...)
			} else {
				tokens = append(tokens, Token{Type: TokenTypeSeparatorOpen, Value: "{"})
				tokens = append(tokens, l.parseText()...)
			}
		case SeparatorClose:
			l.pos++

			tokens = append(tokens, Token{Type: TokenTypeSeparatorClose, Value: "}"})
		case 'm':
			if strings.HasPrefix(string(l.str[l.pos:]), KeywordMatch) {
				tokens = append(tokens, Token{Type: TokenTypeKeyword, Value: KeywordMatch})
				l.pos += len(KeywordMatch)
			}
		case 'w':
			if strings.HasPrefix(string(l.str[l.pos:]), KeywordWhen) {
				tokens = append(tokens, Token{Type: TokenTypeKeyword, Value: KeywordWhen})
				l.pos += len(KeywordWhen)
				tokens = append(tokens, l.parseLiteral())
			}
		}
	}

	return append(tokens, Token{Type: TokenTypeEOF}), nil
}

func (l *lexer) parseFunction() []Token {
	var tokens []Token

	l.pos++

	function := Token{Type: TokenTypeFunction}

	for {
		v := l.current()

		if v == SeparatorClose {
			break
		}

		function.Value += string(v)

		l.pos++
	}

	return append(tokens, function)
}

func (l *lexer) parseText() []Token {
	var tokens []Token
	token := Token{Type: TokenTypeText}

	for {
		if l.current() == SeparatorOpen {
			l.pos++
			tokens = append(tokens, token)

			variable := l.parseVariable()
			tokens = append(tokens, variable...)
			token.Value = string(l.current())
		} else if l.current() == SeparatorClose {
			tokens = append(tokens, token)
			break
		} else {
			token.Value += string(l.current())
		}

		l.pos++
	}

	return tokens
}
func (l *lexer) parseLiteral() Token {
	l.pos++

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

func (l *lexer) parseVariable() []Token {
	tokens := []Token{
		{Type: TokenTypeSeparatorOpen, Value: "{"},
		{Type: TokenTypeVariable},
	}

	variable := &tokens[1]

	l.pos++

	for {
		v := l.current()

		if v == Colon {
			tokens = append(tokens, l.parseFunction()...)
			continue
		}

		if v == SeparatorClose {
			break
		}

		variable.Value += strings.TrimSpace(string(v))

		l.pos++
	}

	l.pos++

	return append(tokens, Token{Type: TokenTypeSeparatorClose, Value: "}"})
}

func Lex(str string) ([]Token, error) {
	l := lexer{str: []rune(str)}

	return l.parse()
}

// combineTextTokens combining Text tokens into one sentence.
func combineTextTokens(tokens, parsedTokens []Token) ([]Token, error) {
	var txt strings.Builder

	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == TokenTypeText {
			if _, err := txt.WriteString(tokens[i].Value); err != nil {
				return []Token{}, errors.New("write Text token")
			}

			if i+1 < len(tokens) && tokens[i+1].Type != TokenTypeText {
				parsedTokens = append(parsedTokens, Token{Type: TokenTypeText, Value: txt.String()})

				txt.Reset()
			}
		} else {
			parsedTokens = append(parsedTokens, tokens[i])
		}
	}

	return parsedTokens, nil
}

// createTokensFromBuffer breaks the input text into tokens that can be processed separately.
func createTokensFromBuffer(buffer []rune, placeholderLevel int) []Token {
	var newTokens []Token

	re := regexp.MustCompile(`\\[^\S\n]?`)
	v := re.ReplaceAllString(strings.TrimSpace(string(buffer)), "")

	if v != "" {
		switch v {
		case KeywordMatch, KeywordLet, KeywordWhen:
			newTokens = append(newTokens, Token{Type: TokenTypeKeyword, Value: v})
		default:
			if placeholderLevel == 0 {
				newTokens = append(newTokens, Token{Type: TokenTypeLiteral, Value: v})
			} else {
				switch buffer[0] {
				case Dollar, Plus, Minus:
					if placeholderLevel > 0 {
						newTokens = append(newTokens,
							Token{
								Type:  TokenTypeVariable,
								Value: strings.ReplaceAll(v, "$", ""),
							})
					}
				case Colon:
					if placeholderLevel > 0 {
						newTokens = append(newTokens, Token{
							Type:  TokenTypeFunction,
							Value: strings.ReplaceAll(v, ":", ""),
						})
					}
				default:
					newTokens = append(newTokens, Token{Type: TokenTypeText, Value: string(buffer)})
				}
			}
		}
	}

	return newTokens
}
