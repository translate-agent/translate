package pot

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type tokenType int

const (
	HeaderLanguage tokenType = iota
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
	Type  tokenType
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

		token, err := parseLine(line, &tokens)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line: %w", err)
		}

		if token == nil {
			continue
		}

		tokens = append(tokens, *token)
	}

	return tokens, nil
}

func parseLine(line string, tokens *[]Token) (*Token, error) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(line, "# Language:"):
		return parseToken(line, HeaderLanguage)
	case strings.HasPrefix(line, "# Plural-Forms:"):
		return parseToken(line, HeaderPluralForms)
	case strings.HasPrefix(line, "# Translator:"):
		return parseToken(line, HeaderTranslator)
	case strings.HasPrefix(line, "msgctxt"):
		return parseToken(line, MsgCtxt)
	case strings.HasPrefix(line, "msgid_plural"):
		return parseToken(line, PluralMsgId)
	case strings.HasPrefix(line, "msgid"):
		return parseToken(line, MsgId)
	case strings.HasPrefix(line, "msgstr["):
		return parsePluralMsgToken(line)
	case strings.HasPrefix(line, "msgstr"):
		return parseToken(line, MsgStr)
	case strings.HasPrefix(line, "#."):
		return parseCommentToken(line, ExtractedComment)
	case strings.HasPrefix(line, "#:"):
		return parseCommentToken(line, Reference)
	case strings.HasPrefix(line, "#| msgid_plural"):
		return parseCommentToken(line, MsgidPluralPrevUntStrPlural)
	case strings.HasPrefix(line, "#| msgid"):
		return parseCommentToken(line, MsgidPrevUntStr)
	case strings.HasPrefix(line, "#| msgctxt"):
		return parseCommentToken(line, MsgctxtPreviousContext)
	case strings.HasPrefix(line, "#,"):
		return parseCommentToken(line, Flag)
	case strings.HasPrefix(line, "# "):
		return parseCommentToken(line, TranslatorComment)
	case strings.HasPrefix(line, `"`):
		return parseMultilineToken(line, tokens)
	default:
		return nil, fmt.Errorf("incorrect format of po tags")
	}
}

func parseToken(line string, tokenType tokenType) (*Token, error) {
	val, err := parseMsgString(line)
	if err != nil {
		return nil, fmt.Errorf("incorrect format of header token: %w", err)
	}

	return &Token{
		Value: val,
		Type:  tokenType,
	}, nil
}

func parsePluralMsgToken(line string) (*Token, error) {
	// Parse the plural index from the line
	indexStart := strings.Index(line, "[") + 1
	indexEnd := strings.Index(line, "]")

	index, err := strconv.Atoi(line[indexStart:indexEnd])
	if err != nil {
		return nil, fmt.Errorf("convert string number to int: %w", err)
	}

	val, err := parseMsgString(line)
	if err != nil {
		return nil, fmt.Errorf("incorrect format of plural msg token: %w", err)
	}

	return &Token{
		Value: val,
		Type:  PluralMsgStr, Index: index,
	}, nil
}

func parseCommentToken(line string, tokenType tokenType) (*Token, error) {
	val, err := parseMsgString(line)
	if err != nil {
		return nil, fmt.Errorf("incorrect format of comment token: %w", err)
	}

	return &Token{
		Value: val,
		Type:  tokenType,
	}, nil
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

func parseMultilineToken(line string, tokens *[]Token) (*Token, error) {
	lastToken := (*tokens)[len(*tokens)-1]
	switch lastToken.Type { //nolint:exhaustive
	case MsgId, PluralMsgId, MsgStr, PluralMsgStr:
		lastToken.Value += " " + parseMultilineString(line)
		lastToken.Value = strings.TrimSpace(lastToken.Value)
		(*tokens)[len(*tokens)-1] = lastToken
	}

	return nil, nil
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
