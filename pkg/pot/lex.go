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
	TokenTypeHeaderLanguage TokenType = iota
	TokenTypeHeaderTranslator
	TokenTypeHeaderPluralForms
	TokenTypeHeaderProjectIDVersion
	TokenTypeHeaderPOTCreationDate
	TokenTypeHeaderPORevisionDate
	TokenTypeHeaderLanguageTeam
	TokenTypeHeaderLastTranslator
	TokenTypeHeaderXGenerator
	TokenTypeHeaderReportMsgidBugsTo
	TokenTypeHeaderMIMEVersion
	TokenTypeHeaderContentType
	TokenTypeHeaderContentTransferEncoding
	TokenTypeMsgCtxt
	TokenTypeMsgID
	TokenTypeMsgStr
	TokenTypePluralMsgID
	TokenTypePluralMsgStr
	TokenTypeTranslatorComment
	TokenTypeExtractedComment
	TokenTypeReference
	TokenTypeFlag
	TokenTypeMsgctxtPreviousContext
	TokenTypeMsgidPluralPrevUntStrPlural
	TokenTypeMsgidPrevUntStr
)

type Token struct {
	Value string
	Type  TokenType
	Index int // plural index for msgstr with plural forms, e.g. msgstr[0], the index is 0
}

func mkToken(tokenType TokenType, value string, opts ...func(t *Token)) Token {
	token := Token{Type: tokenType, Value: value}
	for _, opt := range opts {
		opt(&token)
	}

	return token
}

func withIndex(index int) func(t *Token) {
	return func(t *Token) {
		t.Index = index
	}
}

// Lex function performs lexical analysis on the input by reading lines from the reader
// and parsing each line using the parseLine function.
func Lex(r io.Reader) ([]Token, error) {
	var (
		tokens     []Token
		lineNumber int
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		token, err := parseLine(line, &tokens)
		if err != nil {
			return nil, fmt.Errorf("parse line %d: %w", lineNumber, err)
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
	switch {
	case strings.HasPrefix(line, "\"Language:"):
		return parseToken(line, TokenTypeHeaderLanguage)
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		return parseToken(line, TokenTypeHeaderPluralForms)
	case strings.HasPrefix(line, "\"Translator:"):
		return parseToken(line, TokenTypeHeaderTranslator)
	case strings.HasPrefix(line, "\"Project-Id-Version"):
		return parseToken(line, TokenTypeHeaderProjectIDVersion)
	case strings.HasPrefix(line, "\"POT-Creation-Date"):
		return parseToken(line, TokenTypeHeaderPOTCreationDate)
	case strings.HasPrefix(line, "\"PO-Revision-Date"):
		return parseToken(line, TokenTypeHeaderPORevisionDate)
	case strings.HasPrefix(line, "\"Last-Translator"):
		return parseToken(line, TokenTypeHeaderLastTranslator)
	case strings.HasPrefix(line, "\"Language-Team"):
		return parseToken(line, TokenTypeHeaderLanguageTeam)
	case strings.HasPrefix(line, "\"MIME-Version"):
		return parseToken(line, TokenTypeHeaderMIMEVersion)
	case strings.HasPrefix(line, "\"Content-Type"):
		return parseToken(line, TokenTypeHeaderContentType)
	case strings.HasPrefix(line, "\"Content-Transfer-Encoding"):
		return parseToken(line, TokenTypeHeaderContentTransferEncoding)
	case strings.HasPrefix(line, "\"Report-Msgid-Bugs-To"):
		return parseToken(line, TokenTypeHeaderReportMsgidBugsTo)
	case strings.HasPrefix(line, "\"X-Generator"):
		return parseToken(line, TokenTypeHeaderXGenerator)
	case strings.HasPrefix(line, "msgctxt"):
		return parseToken(line, TokenTypeMsgCtxt)
	case strings.HasPrefix(line, "msgid_plural"):
		return parseToken(line, TokenTypePluralMsgID)
	case strings.HasPrefix(line, "msgid"):
		return parseToken(line, TokenTypeMsgID)
	case strings.HasPrefix(line, "msgstr["):
		return parsePluralMsgToken(line)
	case strings.HasPrefix(line, "msgstr"):
		return parseToken(line, TokenTypeMsgStr)
	case strings.HasPrefix(line, "#."):
		return parseToken(line, TokenTypeExtractedComment)
	case strings.HasPrefix(line, "#:"):
		return parseToken(line, TokenTypeReference)
	case strings.HasPrefix(line, "#| msgid_plural"):
		return parseToken(line, TokenTypeMsgidPluralPrevUntStrPlural)
	case strings.HasPrefix(line, "#| msgid"):
		return parseToken(line, TokenTypeMsgidPrevUntStr)
	case strings.HasPrefix(line, "#| msgctxt"):
		return parseToken(line, TokenTypeMsgctxtPreviousContext)
	case strings.HasPrefix(line, "#,"):
		return parseToken(line, TokenTypeFlag)
	case strings.HasPrefix(line, "#"):
		return parseToken(line, TokenTypeTranslatorComment)
	case strings.HasPrefix(line, `"`):
		return parseMultilineToken(line, tokens)
	default:
		return nil, fmt.Errorf("unknown prefix in '%s'", line)
	}
}

// parseToken function parses the line using the parseMsgString function, which returns a modified string and an error.
// If there is no error a new Token object is created with the parsed value and the specified tokenType.
// TODO: godoc is off, function not needed.
func parseToken(line string, tokenType TokenType) (*Token, error) {
	return &Token{
		Value: parseMsgString(line),
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

	val := parseMsgString(line)

	v := mkToken(TokenTypePluralMsgStr, val, withIndex(index))

	return &v, nil
}

// parseMsgString first checks if the line has a valid prefix using the hasValidPrefix function.
// If the prefix is not valid, it returns an empty string and an error indicating an incorrect format.
func parseMsgString(line string) string {
	n := 2
	fields := strings.SplitN(line, " ", n) // prefix and value, e.g. fields[0] = msgid, fields[1] = "text", etc.

	// No value
	if len(fields) == 1 {
		return ""
	}

	tokenValue := strings.TrimSpace(fields[1])

	// Remove first and last double quotes if they exist
	tokenValue = strings.TrimPrefix(tokenValue, `"`)
	tokenValue = strings.TrimSuffix(tokenValue, `"`)

	return tokenValue
}

// parseMultilineToken function modifies the last token in the slice
// by appending a parsed multiline string from the line. The modified token's value is trimmed of whitespace,
// and the function returns nil for both the token pointer and the error.
func parseMultilineToken(line string, tokens *[]Token) (*Token, error) {
	lastToken := (*tokens)[len(*tokens)-1]
	switch lastToken.Type { //nolint:exhaustive
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypePluralMsgStr:
		lastToken.Value += " " + parseMultilineString(line)
		lastToken.Value = strings.TrimSpace(lastToken.Value)
		(*tokens)[len(*tokens)-1] = lastToken
	}

	return nil, nil //nolint:nilnil
}

// parseMultilineString removes all double quote characters from the input line string and returns the modified string.
func parseMultilineString(line string) string {
	return strings.ReplaceAll(line, `"`, "")
}
