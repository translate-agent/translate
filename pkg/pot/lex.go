package pot

import (
	"bufio"
	"errors"
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
	HeaderProjectIdVersion
	HeaderPOTCreationDate
	HeaderPORevisionDate
	HeaderLanguageTeam
	HeaderLastTranslator
	HeaderXGenerator
	HeaderReportMsgidBugsTo
	HeaderMIMEVersion
	HeaderContentType
	HeaderContentTransferEncoding
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
	"\"Language",
	"\"Plural-Forms",
	"\"Translator",
	"\"Project-Id-Version",
	"\"POT-Creation-Date",
	"\"PO-Revision-Date",
	"\"Last-Translator",
	"\"Language-Team",
	"\"MIME-Version",
	"\"Content-Type",
	"\"Content-Transfer-Encoding",
	"\"X-Generator",
	"\"Report-Msgid-Bugs-To",
}

// Lex function performs lexical analysis on the input by reading lines from the reader
// and parsing each line using the parseLine function.
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

// parseLine function processes the line based on different prefixes
// and returns a pointer to a Token object and an error.
func parseLine(line string, tokens *[]Token) (*Token, error) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(line, "\"Language:"):
		return parseToken(line, HeaderLanguage)
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		return parseToken(line, HeaderPluralForms)
	case strings.HasPrefix(line, "\"Translator:"):
		return parseToken(line, HeaderTranslator)
	case strings.HasPrefix(line, "\"Project-Id-Version"):
		return parseToken(line, HeaderProjectIdVersion)
	case strings.HasPrefix(line, "\"POT-Creation-Date"):
		return parseToken(line, HeaderPOTCreationDate)
	case strings.HasPrefix(line, "\"PO-Revision-Date"):
		return parseToken(line, HeaderPORevisionDate)
	case strings.HasPrefix(line, "\"Last-Translator"):
		return parseToken(line, HeaderLastTranslator)
	case strings.HasPrefix(line, "\"Language-Team"):
		return parseToken(line, HeaderLanguageTeam)
	case strings.HasPrefix(line, "\"MIME-Version"):
		return parseToken(line, HeaderMIMEVersion)
	case strings.HasPrefix(line, "\"Content-Type"):
		return parseToken(line, HeaderContentType)
	case strings.HasPrefix(line, "\"Content-Transfer-Encoding"):
		return parseToken(line, HeaderContentTransferEncoding)
	case strings.HasPrefix(line, "\"Report-Msgid-Bugs-To"):
		return parseToken(line, HeaderReportMsgidBugsTo)
	case strings.HasPrefix(line, "\"X-Generator"):
		return parseToken(line, HeaderXGenerator)
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
		return nil, errors.New("incorrect format of po tags")
	}
}

// parseToken function parses the line using the parseMsgString function, which returns a modified string and an error.
// If there is no error a new Token object is created with the parsed value and the specified tokenType.
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

// parsePluralMsgToken function first parses the plural index from the line string by finding the start and end indices
// of the index value within square brackets. Then converts the extracted substring to an integer.
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

// parseCommentToken function parses the line using the parseMsgString function,
// which returns a modified string and an error.
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

// parseMsgString first checks if the line has a valid prefix using the hasValidPrefix function.
// If the prefix is not valid, it returns an empty string and an error indicating an incorrect format.
func parseMsgString(line string) (string, error) {
	if !hasValidPrefix(line) {
		return "", errors.New("incorrect format of po tags")
	}

	subStrN := 2
	tokenValue := strings.TrimSpace(strings.SplitN(line, " ", subStrN)[1])

	if strings.HasPrefix(tokenValue, `"`) && strings.HasSuffix(tokenValue, `"`) {
		// Remove the quotes and any escaped quotes
		tokenValue = strings.ReplaceAll(tokenValue[1:len(tokenValue)-1], `\"`, `"`)
	} else if strings.HasSuffix(tokenValue, "\\n\"") {
		tokenValue = strings.ReplaceAll(tokenValue[:len(tokenValue)-2], "\\", ``)
		tokenValue = strings.TrimSpace(tokenValue)
	}

	return tokenValue, nil
}

// parseMultilineToken function modifies the last token in the slice
// by appending a parsed multiline string from the line. The modified token's value is trimmed of whitespace,
// and the function returns nil for both the token pointer and the error.
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

// parseMultilineString removes all double quote characters from the input line string and returns the modified string.
func parseMultilineString(line string) string {
	return strings.ReplaceAll(line, `"`, "")
}

// hasValidPrefix determines whether the provided line string has a valid prefix
// by iterating through a collection of valid prefixes.
func hasValidPrefix(line string) bool {
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}
