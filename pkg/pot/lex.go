package pot

import (
	"bufio"
	"errors"
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
	TokenTypeHeaderProjectIdVersion
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
	TokenTypeMsgId
	TokenTypeMsgStr
	TokenTypePluralMsgId
	TokenTypePluralMsgStr
	TokenTypeTranslatorComment
	TokenTypeExtractedComment
	TokenTypeReference
	TokenTypeFlag
	TokenTypeMsgctxtPreviousContext
	TokenTypeMsgidPluralPrevUntStrPlural
	TokenTypeMsgidPrevUntStr
)

func tokenHeaderLanguage(value string) Token {
	return Token{Type: TokenTypeHeaderLanguage, Value: value}
}

func tokenHeaderTranslator(value string) Token {
	return Token{Type: TokenTypeHeaderTranslator, Value: value}
}

func tokenHeaderPluralForms(value string) Token {
	return Token{Type: TokenTypeHeaderPluralForms, Value: value}
}

func tokenHeaderProjectIdVersion(value string) Token {
	return Token{Type: TokenTypeHeaderProjectIdVersion, Value: value}
}

func tokenHeaderPOTCreationDate(value string) Token {
	return Token{Type: TokenTypeHeaderPOTCreationDate, Value: value}
}

func tokenHeaderPORevisionDate(value string) Token {
	return Token{Type: TokenTypeHeaderPORevisionDate, Value: value}
}

func tokenHeaderLanguageTeam(value string) Token {
	return Token{Type: TokenTypeHeaderLanguageTeam, Value: value}
}

func tokenHeaderLastTranslator(value string) Token {
	return Token{Type: TokenTypeHeaderLastTranslator, Value: value}
}

func tokenHeaderXGenerator(value string) Token {
	return Token{Type: TokenTypeHeaderXGenerator, Value: value}
}

func tokenHeaderReportMsgidBugsTo(value string) Token {
	return Token{Type: TokenTypeHeaderReportMsgidBugsTo, Value: value}
}

func tokenHeaderMIMEVersion(value string) Token {
	return Token{Type: TokenTypeHeaderMIMEVersion, Value: value}
}

func tokenHeaderContentType(value string) Token {
	return Token{Type: TokenTypeHeaderContentType, Value: value}
}

func tokenHeaderContentTransferEncoding(value string) Token {
	return Token{Type: TokenTypeHeaderContentTransferEncoding, Value: value}
}

func tokenMsgCtxt(value string) Token {
	return Token{Type: TokenTypeMsgCtxt, Value: value}
}

func tokenMsgId(value string) Token {
	return Token{Type: TokenTypeMsgId, Value: value}
}

func tokenMsgStr(value string) Token {
	return Token{Type: TokenTypeMsgStr, Value: value}
}

func tokenPluralMsgId(value string) Token {
	return Token{Type: TokenTypePluralMsgId, Value: value}
}

func tokenPluralMsgStr(value string, index int) Token {
	return Token{Type: TokenTypePluralMsgStr, Value: value, Index: index}
}

func tokenTranslatorComment(value string) Token {
	return Token{Type: TokenTypeTranslatorComment, Value: value}
}

func tokenExtractedComment(value string) Token {
	return Token{Type: TokenTypeExtractedComment, Value: value}
}

func tokenReference(value string) Token {
	return Token{Type: TokenTypeReference, Value: value}
}

func tokenFlag(value string) Token {
	return Token{Type: TokenTypeFlag, Value: value}
}

func tokenMsgctxtPreviousContext(value string) Token {
	return Token{Type: TokenTypeMsgctxtPreviousContext, Value: value}
}

func tokenMsgidPluralPrevUntStrPlural(value string) Token {
	return Token{Type: TokenTypeMsgidPluralPrevUntStrPlural, Value: value}
}

func tokenMsgidPrevUntStr(value string) Token {
	return Token{Type: TokenTypeMsgidPrevUntStr, Value: value}
}

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
	"#",
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
	var (
		tokens     []Token
		lineNumber int
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()

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
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(line, "\"Language:"):
		return parseToken(line, TokenTypeHeaderLanguage)
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		return parseToken(line, TokenTypeHeaderPluralForms)
	case strings.HasPrefix(line, "\"Translator:"):
		return parseToken(line, TokenTypeHeaderTranslator)
	case strings.HasPrefix(line, "\"Project-Id-Version"):
		return parseToken(line, TokenTypeHeaderProjectIdVersion)
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
		return parseToken(line, TokenTypePluralMsgId)
	case strings.HasPrefix(line, "msgid"):
		return parseToken(line, TokenTypeMsgId)
	case strings.HasPrefix(line, "msgstr["):
		return parsePluralMsgToken(line)
	case strings.HasPrefix(line, "msgstr"):
		return parseToken(line, TokenTypeMsgStr)
	case strings.HasPrefix(line, "#."):
		return parseCommentToken(line, TokenTypeExtractedComment)
	case strings.HasPrefix(line, "#:"):
		return parseCommentToken(line, TokenTypeReference)
	case strings.HasPrefix(line, "#| msgid_plural"):
		return parseCommentToken(line, TokenTypeMsgidPluralPrevUntStrPlural)
	case strings.HasPrefix(line, "#| msgid"):
		return parseCommentToken(line, TokenTypeMsgidPrevUntStr)
	case strings.HasPrefix(line, "#| msgctxt"):
		return parseCommentToken(line, TokenTypeMsgctxtPreviousContext)
	case strings.HasPrefix(line, "#,"):
		return parseCommentToken(line, TokenTypeFlag)
	case strings.HasPrefix(line, "#"):
		return parseCommentToken(line, TokenTypeTranslatorComment)
	case strings.HasPrefix(line, `"`):
		return parseMultilineToken(line, tokens)
	default:
		return nil, fmt.Errorf("unknown prefix in '%s'", line)
	}
}

// parseToken function parses the line using the parseMsgString function, which returns a modified string and an error.
// If there is no error a new Token object is created with the parsed value and the specified tokenType.
func parseToken(line string, tokenType TokenType) (*Token, error) {
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

	v := tokenPluralMsgStr(val, index)

	return &v, nil
}

// parseCommentToken function parses the line using the parseMsgString function,
// which returns a modified string and an error.
func parseCommentToken(line string, tokenType TokenType) (*Token, error) {
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

	n := 2
	fields := strings.SplitN(line, " ", n)

	var tokenValue string

	if len(fields) == n {
		tokenValue = strings.TrimSpace(fields[1])
	}

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
	case TokenTypeMsgId, TokenTypePluralMsgId, TokenTypeMsgStr, TokenTypePluralMsgStr:
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
