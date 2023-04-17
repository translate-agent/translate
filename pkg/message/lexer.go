package message

import (
	"errors"
	"strings"
)

type Token struct {
	Value string
	Level int
	Type  TokenType
}

type TokenType int

const (
	TokenTypeUnknown TokenType = iota
	TokenTypeKeyword
	TokenTypePlaceholderOpen
	TokenTypePlaceholderClose
	TokenTypeLiteral
	TokenTypeText
	TokenTypeFunction
	TokenTypeVariable
)

const (
	Match  = "match"
	Let    = "let"
	When   = "when"
	Dollar = '$'
	Colon  = ':'
	Plus   = '+'
	Minus  = '-'
)

func Lex(str string) ([]Token, error) {
	var (
		tokens           []Token
		parsedTokens     []Token
		runes            []rune
		placeholderLevel int
	)

	for i, r := range str {
		switch r {
		case ' ':
			if len(runes) > 0 {
				runes = append(runes, r)
				tokens = append(tokens, createTokensFromBuffer(runes, placeholderLevel)...)

				runes = []rune{}
			} else {
				runes = append(runes, r)
			}
		case '{':
			if len(runes) > 0 {
				tokens = append(tokens, createTokensFromBuffer(runes, placeholderLevel)...)

				runes = []rune{}
			}

			placeholderLevel++

			tokens = append(tokens, Token{Type: TokenTypePlaceholderOpen, Value: "{", Level: placeholderLevel})
		case '}':
			if len(runes) > 0 {
				tokens = append(tokens, createTokensFromBuffer(runes, placeholderLevel)...)

				runes = []rune{}
			}

			tokens = append(tokens, Token{Type: TokenTypePlaceholderClose, Value: "}", Level: placeholderLevel})
			placeholderLevel--
		case '$', ':', '+', '-':
			if i+1 < len(str) && str[i+1] == ' ' {
				return []Token{}, errors.New("variable or function name starts with a space")
			}

			if len(runes) > 0 {
				tokens = append(tokens, createTokensFromBuffer(runes, placeholderLevel)...)
				runes = []rune{}
			}

			runes = append(runes, r)
		default:
			runes = append(runes, r)
		}
	}

	if len(runes) > 0 {
		tokens = append(tokens, createTokensFromBuffer(runes, placeholderLevel)...)
	}

	parsedTokens, err := combineTextTokens(tokens, parsedTokens)
	if err != nil {
		return []Token{}, errors.New("combine Text tokens")
	}

	return parsedTokens, nil
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
				parsedTokens = append(parsedTokens, Token{Type: TokenTypeText, Value: txt.String(), Level: tokens[i].Level})

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

	v := strings.TrimSpace(string(buffer))
	switch v {
	case Match, Let, When:
		newTokens = append(newTokens, Token{Type: TokenTypeKeyword, Value: v, Level: placeholderLevel})
	default:
		if placeholderLevel == 0 {
			newTokens = append(newTokens, Token{Type: TokenTypeLiteral, Value: v, Level: placeholderLevel})
		} else {
			switch buffer[0] {
			case Dollar, Plus, Minus:
				if placeholderLevel > 0 {
					newTokens = append(newTokens,
						Token{
							Type:  TokenTypeVariable,
							Value: strings.ReplaceAll(v, "$", ""),
							Level: placeholderLevel,
						})
				}
			case Colon:
				if placeholderLevel > 0 {
					newTokens = append(newTokens, Token{
						Type:  TokenTypeFunction,
						Value: strings.ReplaceAll(v, ":", ""),
						Level: placeholderLevel,
					})
				}
			default:
				newTokens = append(newTokens, Token{Type: TokenTypeText, Value: string(buffer), Level: placeholderLevel})
			}
		}
	}

	return newTokens
}
