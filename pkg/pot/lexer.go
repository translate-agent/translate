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
	MsgidPrevUntStr
)

type Token struct {
	Value string
	Type  TokenType
	Index int // plural index for msgstr with plural forms
}

var validPrefixes = []string{
	"msgctxt ",
	"msgid ",
	"msgstr ",
	"msgid_plural ",
	"msgstr[0] ",
	"msgstr[1] ",
	"msgstr[2] ",
	"msgstr[3] ",
	"msgstr[4] ",
	"msgstr[5] ",
	"#: ",
	"# ",
	"#. ",
	"#, ",
	"#| msgctxt ",
	"#| msgid_plural ",
	"#| msgid ",
}

func Lex(r io.Reader) ([]Token, error) {
	var tokens []Token

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if err := parseLine(line, &tokens); err != nil {
			return nil, err
		}
	}

	return tokens, nil
}

func parseLine(line string, tokens *[]Token) error {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	switch {
	case strings.HasPrefix(line, "# Language:"):
		return parseHeaderToken(line, HeaderLanguage, tokens)
	case strings.HasPrefix(line, "# Plural-Forms:"):
		return parseHeaderToken(line, HeaderPluralForms, tokens)
	case strings.HasPrefix(line, "# Translator:"):
		return parseHeaderToken(line, HeaderTranslator, tokens)
	case strings.HasPrefix(line, "msgctxt"):
		return parseMsgToken(line, MsgCtxt, tokens)
	case strings.HasPrefix(line, "msgid_plural"):
		return parseMsgToken(line, PluralMsgId, tokens)
	case strings.HasPrefix(line, "msgid"):
		return parseMsgToken(line, MsgId, tokens)
	case strings.HasPrefix(line, "msgstr["):
		return parsePluralMsgToken(line, tokens)
	case strings.HasPrefix(line, "msgstr"):
		return parseMsgToken(line, MsgStr, tokens)
	case strings.HasPrefix(line, "#."):
		return parseCommentToken(line, ExtractedComment, tokens)
	case strings.HasPrefix(line, "#:"):
		return parseCommentToken(line, Reference, tokens)
	case strings.HasPrefix(line, "#| msgid_plural"):
		return parseCommentToken(line, MsgidPluralPrevUntStrPlural, tokens)
	case strings.HasPrefix(line, "#| msgid"):
		return parseCommentToken(line, MsgidPrevUntStr, tokens)
	case strings.HasPrefix(line, "#| msgctxt"):
		return parseCommentToken(line, MsgctxtPreviousContext, tokens)
	case strings.HasPrefix(line, "#,"):
		return parseCommentToken(line, Flag, tokens)
	case strings.HasPrefix(line, "# "):
		return parseCommentToken(line, TranslatorComment, tokens)
	case strings.HasPrefix(line, `"`):
		return parseMultilineToken(line, tokens)
	default:
		return fmt.Errorf("incorrect format of po tags")
	}
}

func parseHeaderToken(line string, tokenType TokenType, tokens *[]Token) error {
	val, err := parseMsgString(line)
	if err != nil {
		return fmt.Errorf("incorrect format of header token: %w", err)
	}

	*tokens = append(*tokens, Token{Value: val, Type: tokenType})

	return nil
}

func parseMsgToken(line string, tokenType TokenType, tokens *[]Token) error {
	val, err := parseMsgString(line)
	if err != nil {
		return fmt.Errorf("incorrect format of msg token: %w", err)
	}

	*tokens = append(*tokens, Token{Value: val, Type: tokenType})

	return nil
}

func parsePluralMsgToken(line string, tokens *[]Token) error {
	// Parse the plural index from the line
	indexStart := strings.Index(line, "[") + 1
	indexEnd := strings.Index(line, "]")

	index, err := strconv.Atoi(line[indexStart:indexEnd])
	if err != nil {
		return fmt.Errorf("convert string number to int: %w", err)
	}

	val, err := parseMsgString(line)
	if err != nil {
		return fmt.Errorf("incorrect format of plural msg token: %w", err)
	}

	*tokens = append(*tokens, Token{Value: val, Type: PluralMsgStr, Index: index})

	return nil
}

func parseCommentToken(line string, tokenType TokenType, tokens *[]Token) error {
	val, err := parseMsgString(line)
	if err != nil {
		return fmt.Errorf("incorrect format of comment token: %w", err)
	}

	*tokens = append(*tokens, Token{Value: val, Type: tokenType})

	return nil
}

func parseMsgString(line string) (string, error) {
	if !hasValidPrefix(line) {
		return "", fmt.Errorf("incorrect format of po tags")
	}

	subStrN := 2
	tokenValue := strings.TrimSpace(strings.SplitN(line, " ", subStrN)[1])

	if strings.HasPrefix(tokenValue, `"`) && strings.HasSuffix(tokenValue, `"`) {
		// Remove the quotes and any escaped quotes
		tokenValue = strings.ReplaceAll(tokenValue[1:len(tokenValue)-1], `\"`, `"`)
	}

	return tokenValue, nil
}

func parseMultilineToken(line string, tokens *[]Token) error {
	tokenToAppend := (*tokens)[len(*tokens)-1]
	switch tokenToAppend.Type { //nolint:exhaustive
	case MsgId, PluralMsgId, MsgStr, PluralMsgStr:
		tokenToAppend.Value += " " + parseMultilineString(line)
		tokenToAppend.Value = strings.TrimSpace(tokenToAppend.Value)
		(*tokens)[len(*tokens)-1] = tokenToAppend
	}

	return nil
}

func parseMultilineString(line string) string {
	return strings.ReplaceAll(line, `"`, "")
}

func hasValidPrefix(line string) bool {
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}
