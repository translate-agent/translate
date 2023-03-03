package pot

import (
	"bufio"
	"fmt"
	"io"
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

func Lex(r *bufio.Reader) ([]Token, error) {
	var tokens []Token

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("reading line: %w", err)
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# Language:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: HeaderLanguage})
		case strings.HasPrefix(line, "# Plural-Forms:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: HeaderPluralForms})
		case strings.HasPrefix(line, "# Translator:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: HeaderTranslator})
		case strings.HasPrefix(line, "msgctxt"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgCtxt})
		case strings.HasPrefix(line, "msgid_plural"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: PluralMsgId})
		case strings.HasPrefix(line, "msgid"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgId})
		case strings.HasPrefix(line, "msgstr["):
			// Parse the plural index from the line
			indexStart := strings.Index(line, "[") + 1
			indexEnd := strings.Index(line, "]")

			index, err := strconv.Atoi(line[indexStart:indexEnd])
			if err != nil {
				continue
			}

			tokens = append(tokens, Token{parseMsgString(line), PluralMsgStr, index})
		case strings.HasPrefix(line, "msgstr"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgStr})
		case strings.HasPrefix(line, "#."):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: ExtractedComment})
		case strings.HasPrefix(line, "#:"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: Reference})
		case strings.HasPrefix(line, "#| msgid_plural"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgidPluralPrevUntStrPlural})
		case strings.HasPrefix(line, "#| msgid"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgidPrevUntdStr})
		case strings.HasPrefix(line, "#| msgctxt"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: MsgctxtPreviousContext})
		case strings.HasPrefix(line, "#,"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: Flag})
		case strings.HasPrefix(line, "#"):
			tokens = append(tokens, Token{Value: parseMsgString(line), Type: TranslatorComment})
		case strings.HasPrefix(line, `"`):
			tokenToAppend := tokens[len(tokens)-1]
			switch tokenToAppend.Type { //nolint:exhaustive
			case MsgId, PluralMsgId, MsgStr, PluralMsgStr:
				tokenToAppend.Value += " " + parseMultilineString(line)
				tokenToAppend.Value = strings.TrimSpace(tokenToAppend.Value)
				tokens[len(tokens)-1] = tokenToAppend
			}
		}
	}

	return tokens, nil
}

func parseMsgString(line string) string {
	subStrN := 2
	tokenValue := strings.TrimSpace(strings.SplitN(line, " ", subStrN)[1])

	return strings.ReplaceAll(tokenValue, `"`, "")
}

func parseMultilineString(line string) string {
	return strings.ReplaceAll(line, `"`, "")
}
