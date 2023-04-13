package messageFormat

import (
	"fmt"
	"strings"

	_ "golang.org/x/exp/slices"
)

type Token struct {
	Value string
	Level int
	Type  tokenType
}

type tokenType int

const (
	Keyword          tokenType = iota
	PlaceholderOpen  tokenType = iota
	PlaceholderClose tokenType = iota
	Literal          tokenType = iota
	Text             tokenType = iota
	Function
	Variable
)

func Lex(str string) ([]Token, error) {
	var (
		tokens           []Token
		parsedTokens     []Token
		buf              []rune
		placeholderLevel int
	)

	for i, r := range str {
		switch r {
		case ' ':
			if len(buf) > 0 {
				buf = append(buf, r)
				tokenizeBuffer(buf, &tokens, &placeholderLevel)
				buf = []rune{}
			} else {
				buf = append(buf, r)
			}
		case '{':
			if len(buf) > 0 {
				tokenizeBuffer(buf, &tokens, &placeholderLevel)
				buf = []rune{}
			}

			placeholderLevel++

			tokens = append(tokens, Token{Type: PlaceholderOpen, Value: "{", Level: placeholderLevel})
		case '}':
			if len(buf) > 0 {
				tokenizeBuffer(buf, &tokens, &placeholderLevel)
				buf = []rune{}
			}

			tokens = append(tokens, Token{Type: PlaceholderClose, Value: "}", Level: placeholderLevel})
			placeholderLevel--
		case '$', ':', '+', '-':
			if str[i+1] == ' ' {
				return []Token{}, fmt.Errorf("variable or function name starts with a space")
			}

			if len(buf) > 0 {
				tokenizeBuffer(buf, &tokens, &placeholderLevel)
				buf = []rune{}
			}

			buf = append(buf, r)
		default:
			buf = append(buf, r)
		}
	}

	if len(buf) > 0 {
		tokenizeBuffer(buf, &tokens, &placeholderLevel)
	}

	parsedTokens = combineTextTokens(tokens, parsedTokens)

	return parsedTokens, nil
}

func combineTextTokens(tokens []Token, parsedTokens []Token) []Token {
	txt := ""

	for y := range tokens {
		if tokens[y].Type == Text {
			txt += tokens[y].Value

			if tokens[y+1].Type != Text {
				parsedTokens = append(parsedTokens, Token{Type: Text, Value: txt, Level: tokens[y].Level})
				txt = ""
			}
		} else {
			parsedTokens = append(parsedTokens, tokens[y])
		}
	}

	return parsedTokens
}

func tokenizeBuffer(buf []rune, tokens *[]Token, placeholderLevel *int) {
	switch strings.TrimSpace(string(buf)) {
	case "match", "let", "when":
		*tokens = append(*tokens, Token{Type: Keyword, Value: string(buf), Level: *placeholderLevel})
	default:
		if *placeholderLevel == 0 {
			*tokens = append(*tokens, Token{Type: Literal, Value: string(buf), Level: *placeholderLevel})
		} else {
			switch buf[0] {
			case '$', '+', '-':
				if *placeholderLevel > 0 {
					str := strings.ReplaceAll(string(buf), "$", "")
					*tokens = append(*tokens, Token{Type: Variable, Value: str, Level: *placeholderLevel})
				}
			case ':':
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
