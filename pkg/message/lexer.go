package message

import (
	"errors"
	"strings"
)

type Token struct {
	Value string
	Level int
	Type  tokenType
}

type tokenType int

const (
	Keyword tokenType = iota
	PlaceholderOpen
	PlaceholderClose
	Literal
	Text
	Function
	Variable
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
				tokenizeBuffer(runes, &tokens, &placeholderLevel)
				runes = []rune{}
			} else {
				runes = append(runes, r)
			}
		case '{':
			if len(runes) > 0 {
				tokenizeBuffer(runes, &tokens, &placeholderLevel)
				runes = []rune{}
			}

			placeholderLevel++

			tokens = append(tokens, Token{Type: PlaceholderOpen, Value: "{", Level: placeholderLevel})
		case '}':
			if len(runes) > 0 {
				tokenizeBuffer(runes, &tokens, &placeholderLevel)
				runes = []rune{}
			}

			tokens = append(tokens, Token{Type: PlaceholderClose, Value: "}", Level: placeholderLevel})
			placeholderLevel--
		case '$', ':', '+', '-':
			if i+1 < len(str) && str[i+1] == ' ' {
				return []Token{}, errors.New("variable or function name starts with a space")
			}

			if len(runes) > 0 {
				tokenizeBuffer(runes, &tokens, &placeholderLevel)
				runes = []rune{}
			}

			runes = append(runes, r)
		default:
			runes = append(runes, r)
		}
	}

	if len(runes) > 0 {
		tokenizeBuffer(runes, &tokens, &placeholderLevel)
	}

	parsedTokens = combineTextTokens(tokens, parsedTokens)

	return parsedTokens, nil
}

func combineTextTokens(tokens, parsedTokens []Token) []Token {
	txt := ""

	for i := range tokens {
		if tokens[i].Type == Text {
			txt += tokens[i].Value

			if tokens[i+1].Type != Text {
				parsedTokens = append(parsedTokens, Token{Type: Text, Value: txt, Level: tokens[i].Level})
				txt = ""
			}
		} else {
			parsedTokens = append(parsedTokens, tokens[i])
		}
	}

	return parsedTokens
}

func tokenizeBuffer(buf []rune, tokens *[]Token, placeholderLevel *int) {
	switch strings.TrimSpace(string(buf)) {
	case Match, Let, When:
		*tokens = append(*tokens, Token{Type: Keyword, Value: string(buf), Level: *placeholderLevel})
	default:
		if *placeholderLevel == 0 {
			*tokens = append(*tokens, Token{Type: Literal, Value: string(buf), Level: *placeholderLevel})
		} else {
			switch buf[0] {
			case Dollar, Plus, Minus:
				if *placeholderLevel > 0 {
					str := strings.ReplaceAll(string(buf), "$", "")
					*tokens = append(*tokens, Token{Type: Variable, Value: str, Level: *placeholderLevel})
				}
			case Colon:
				if *placeholderLevel > 0 {
					str := strings.ReplaceAll(string(buf), ":", "")
					*tokens = append(*tokens, Token{Type: Function, Value: str, Level: *placeholderLevel})
				}
			default:
				*tokens = append(*tokens, Token{Type: Text, Value: string(buf), Level: *placeholderLevel})
			}
		}
	}
}
