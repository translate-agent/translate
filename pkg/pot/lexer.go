package pot

import (
	"bufio"
	"strconv"
	"strings"
)

type TokenType int

const (
	HeaderLanguage TokenType = iota
	HeaderTranslator
	HeaderPluralForms
	MsgCtxt
	MsgId
	MsgStr
	PluralMsgId
	PluralMsgStr
	TranslatorComment
	ExtractedComment
	Reference
	Flag
	MsgctxtPreviousContext
	MsgidPluralPrevUntStrPlural
	MsgidPrevUntdStr
)

type Token struct {
	Value string
	Type  TokenType
	Index int // plural index for msgstr with plural forms
}

func Lex(r *bufio.Reader) []Token {
	var tokens []Token

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# Language:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: HeaderLanguage})
			continue
		case strings.HasPrefix(line, "# Plural-Forms:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: HeaderPluralForms})
			continue
		case strings.HasPrefix(line, "# Translator:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: HeaderTranslator})
			continue
		case strings.HasPrefix(line, "msgctxt"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgCtxt})
			continue
		case strings.HasPrefix(line, "msgid_plural"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: PluralMsgId})
			continue
		case strings.HasPrefix(line, "msgid"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgId})
			continue
		case strings.HasPrefix(line, "msgstr["):
			// Parse the plural index from the line
			indexStart := strings.Index(line, "[") + 1
			indexEnd := strings.Index(line, "]")

			index, err := strconv.Atoi(line[indexStart:indexEnd])
			if err != nil {
				continue
			}

			tokens = append(tokens, Token{parseMsgString(line), PluralMsgStr, index})

			continue
		case strings.HasPrefix(line, "msgstr"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgStr})
			continue
		case strings.HasPrefix(line, "#."):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: ExtractedComment})
			continue
		case strings.HasPrefix(line, "#:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: Reference})
			continue
		case strings.HasPrefix(line, "#| msgid_plural"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgidPluralPrevUntStrPlural})
			continue
		case strings.HasPrefix(line, "#| msgid"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgidPrevUntdStr})
			continue
		case strings.HasPrefix(line, "#| msgctxt"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgctxtPreviousContext})
			continue
		case strings.HasPrefix(line, "#,"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: Flag})
			continue
		case strings.HasPrefix(line, "#"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: TranslatorComment})
			continue
		default:
			if strings.HasPrefix(line, `"`) {
				tokenToAppend := tokens[len(tokens)-1]
				if tokenToAppend.Type == MsgId || tokenToAppend.Type == MsgStr ||
					tokenToAppend.Type == PluralMsgId || tokenToAppend.Type == PluralMsgStr {
					tokenToAppend.Value += " " + parseMultilineString(line)
					tokenToAppend.Value = strings.TrimSpace(tokenToAppend.Value)
					tokens[len(tokens)-1] = tokenToAppend
				}
			}

			continue
		}
	}

	return tokens
}

func parseMsgString(line string) string {
	subStrN := 2
	tokenValue := strings.TrimSpace(strings.SplitN(line, " ", subStrN)[1])

	return strings.ReplaceAll(tokenValue, `"`, "")
}

func parseMultilineString(line string) string {
	return strings.ReplaceAll(line, `"`, "")
}
