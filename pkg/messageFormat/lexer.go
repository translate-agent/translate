package messageFormat

import (
	_ "golang.org/x/exp/slices"
	"strings"
)

type Token struct {
	Type  tokenType
	Value string
	Level int
}

type tokenType int

const (
	keyword          tokenType = iota
	placeholderOpen  tokenType = iota
	placeholderClose tokenType = iota
	literal          tokenType = iota
	text             tokenType = iota
	function
	variable
)

func tokenize(str string) []Token {
	var tokens []Token
	var buf []rune
	var placeholderLevel int

	for _, r := range str {
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
			tokens = append(tokens, Token{Type: placeholderOpen, Value: "{"})
			placeholderLevel++
		case '}':
			if len(buf) > 0 {
				tokenizeBuffer(buf, &tokens, &placeholderLevel)
				buf = []rune{}
			}
			tokens = append(tokens, Token{Type: placeholderClose, Value: "}"})
			placeholderLevel--
		case '$', ':', '+', '-':
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

	var tokens2 []Token
	txt := ""
	for y, _ := range tokens {

		if tokens[y].Type == text {
			txt += tokens[y].Value
			if tokens[y+1].Type != text {
				tokens2 = append(tokens2, Token{Type: text, Value: txt})
				txt = ""
			}
		} else {
			tokens2 = append(tokens2, tokens[y])
		}
	}

	return tokens2

}

func tokenizeBuffer(buf []rune, tokens *[]Token, placeholderLevel *int) {
	strNoSpace := strings.TrimSpace(string(buf))

	switch strNoSpace {
	case "match", "let", "when":
		*tokens = append(*tokens, Token{Type: keyword, Value: string(buf)})
	case "$", "+", "-":
		str := strings.ReplaceAll(string(buf), "$", "")
		if *placeholderLevel > 0 {
			*tokens = append(*tokens, Token{Type: variable, Value: str})
		} else {
			*tokens = append(*tokens, Token{Type: text, Value: string(buf)})
		}
	default:
		if *placeholderLevel == 0 {
			*tokens = append(*tokens, Token{Type: literal, Value: string(buf)})
		} else {
			if (buf[0] == '$' || buf[0] == '+' || buf[0] == '-') && *placeholderLevel > 0 {
				str := strings.ReplaceAll(string(buf), "$", "")
				*tokens = append(*tokens, Token{Type: variable, Value: str})
			} else if (buf[0] == ':' || buf[0] == '+' || buf[0] == '-') && *placeholderLevel > 0 {
				str := strings.ReplaceAll(string(buf), ":", "")
				*tokens = append(*tokens, Token{Type: function, Value: str})
			} else {
				*tokens = append(*tokens, Token{Type: text, Value: string(buf)})
			}
		}
	}
}
