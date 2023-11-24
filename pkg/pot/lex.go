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
	// File header tokens.
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
	TokenTypeGeneratedBy
	// Message tokens.
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

		token, err := parseLine(line, tokens)
		if err != nil {
			return nil, fmt.Errorf("parse line \"%s\": %w", line, err)
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
func parseLine(line string, tokens []Token) (*Token, error) {
	var token TokenType

	switch {
	default:
		return nil, fmt.Errorf("unknown prefix in '%s'", line)
	case strings.HasPrefix(line, "\"Language:"):
		token = TokenTypeHeaderLanguage
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		token = TokenTypeHeaderPluralForms
	case strings.HasPrefix(line, "\"Translator:"):
		token = TokenTypeHeaderTranslator
	case strings.HasPrefix(line, "\"Project-Id-Version"):
		token = TokenTypeHeaderProjectIDVersion
	case strings.HasPrefix(line, "\"POT-Creation-Date"):
		token = TokenTypeHeaderPOTCreationDate
	case strings.HasPrefix(line, "\"PO-Revision-Date"):
		token = TokenTypeHeaderPORevisionDate
	case strings.HasPrefix(line, "\"Last-Translator"):
		token = TokenTypeHeaderLastTranslator
	case strings.HasPrefix(line, "\"Language-Team"):
		token = TokenTypeHeaderLanguageTeam
	case strings.HasPrefix(line, "\"MIME-Version"):
		token = TokenTypeHeaderMIMEVersion
	case strings.HasPrefix(line, "\"Content-Type"):
		token = TokenTypeHeaderContentType
	case strings.HasPrefix(line, "\"Content-Transfer-Encoding"):
		token = TokenTypeHeaderContentTransferEncoding
	case strings.HasPrefix(line, "\"Report-Msgid-Bugs-To"):
		token = TokenTypeHeaderReportMsgidBugsTo
	case strings.HasPrefix(line, "\"X-Generator"):
		token = TokenTypeHeaderXGenerator
	case strings.HasPrefix(line, "\"Generated-By"):
		token = TokenTypeGeneratedBy
	case strings.HasPrefix(line, "msgctxt"):
		token = TokenTypeMsgCtxt
	case strings.HasPrefix(line, "msgid_plural"):
		token = TokenTypePluralMsgID
	case strings.HasPrefix(line, "msgid"):
		token = TokenTypeMsgID
	case strings.HasPrefix(line, "msgstr["):
		token = TokenTypePluralMsgStr
	case strings.HasPrefix(line, "msgstr"):
		token = TokenTypeMsgStr
	case strings.HasPrefix(line, "#."):
		token = TokenTypeExtractedComment
	case strings.HasPrefix(line, "#:"):
		token = TokenTypeReference
	case strings.HasPrefix(line, "#| msgid_plural"):
		token = TokenTypeMsgidPluralPrevUntStrPlural
	case strings.HasPrefix(line, "#| msgid"):
		token = TokenTypeMsgidPrevUntStr
	case strings.HasPrefix(line, "#| msgctxt"):
		token = TokenTypeMsgctxtPreviousContext
	case strings.HasPrefix(line, "#,"):
		token = TokenTypeFlag
	case strings.HasPrefix(line, "#"):
		token = TokenTypeTranslatorComment
	// Special case for multiline strings, i.e. msgid, msgstr, msgid_plural, msgstr[*]
	// When we encounter header token which is not defined in the TokenType enum, it tries parse the header
	// as a multiline string. As workaround, we allow parsing it as multiline string if last token was not header token.
	case strings.HasPrefix(line, `"`) && tokens[len(tokens)-1].Type >= TokenTypeMsgCtxt:
		return nil, parseMultilineValue(line, tokens)
	}

	return parseToken(line, token)
}

// parseToken function parses the value of the token based on the token type.
// On PluralMsgStr token type, it also tries parses the plural index.
func parseToken(line string, tokenType TokenType) (*Token, error) {
	value, err := parseValue(tokenType, line)
	if err != nil {
		return nil, fmt.Errorf("trim quotes: %w", err)
	}

	token := Token{Type: tokenType, Value: value}

	// We assume that headers token have a newline at the end so we must trim it.
	if tokenType <= TokenTypeGeneratedBy {
		token.Value = strings.TrimSuffix(token.Value, "\\n")
	}

	if tokenType != TokenTypePluralMsgStr {
		return &token, nil
	}

	// Parse the plural index from the line
	indexStart := strings.Index(line, "[")
	indexEnd := strings.Index(line, "]")

	if indexStart == -1 || indexEnd == -1 || indexStart >= indexEnd {
		return nil, errors.New("improperly formatted msgstr[*] or msgstr[*] not found in the string")
	}

	indexStr := line[indexStart+1 : indexEnd]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("convert string '%s' to int: %w", indexStr, err)
	}

	token.Index = index

	return &token, nil
}

// parseValue parses the value of the token.
// e.g.
//
//	"Language: en-US\n" -> en-US\n
//	msgid "some text" -> some text
//	#, python-format -> python-format
//	#: superset/key_value/exceptions.py:54 -> superset/key_value/exceptions.py:54
func parseValue(tokenType TokenType, line string) (string, error) {
	n := 2
	fields := strings.SplitN(line, " ", n) // prefix and value, e.g. fields[0] = msgid, fields[1] = "text", etc.

	// No value, only prefix
	if len(fields) != n {
		return "", nil
	}

	return trimQuotes(tokenType, fields[1])
}

// parseMultilineValue parses the value of the multiline msgid, msgstr, msgid_plural, msgstr[*] tokens.
// Instead of returning a parsed value, it appends it to the last token in the tokens slice.
func parseMultilineValue(line string, tokens []Token) error {
	lastToken := &tokens[len(tokens)-1]

	switch lastToken.Type { //nolint:exhaustive
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypePluralMsgStr:
		value, err := trimQuotes(lastToken.Type, line)
		if err != nil {
			return fmt.Errorf("trim quotes: %w", err)
		}

		lastToken.Value += "\n" + value
	default:
		return fmt.Errorf("unsupported multiline string for token type: '%d'", lastToken.Type)
	}

	return nil
}

// trimQuotes tries to trim the quotes from the token value based on the token type.
// It returns an error if the token value is not properly formatted.
// e.g.
//
// values for msgid, msgstr, msgid_plural, msgstr[*]
//
//	"text1" -> text1
//	text2 -> error
//
// values for headers
//
//	en-US\n" -> en-US\n
//	lv-LV\n -> error
func trimQuotes(tokenType TokenType, value string) (string, error) {
	switch tokenType {
	// Quotes at the beginning and end of the string.
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypePluralMsgStr, TokenTypeMsgCtxt:
		if value[0] != '"' || value[len(value)-1] != '"' {
			return "", fmt.Errorf("token value %d must start and end with quote", tokenType)
		}

		value = strings.Trim(value, `"`)

	// Quotes at the end of the string.
	case TokenTypeHeaderLanguage, TokenTypeHeaderTranslator, TokenTypeHeaderPluralForms,
		TokenTypeHeaderProjectIDVersion, TokenTypeHeaderPOTCreationDate, TokenTypeHeaderPORevisionDate,
		TokenTypeHeaderLanguageTeam, TokenTypeHeaderLastTranslator, TokenTypeHeaderXGenerator,
		TokenTypeHeaderReportMsgidBugsTo, TokenTypeHeaderMIMEVersion, TokenTypeHeaderContentType,
		TokenTypeHeaderContentTransferEncoding, TokenTypeGeneratedBy:
		if value[len(value)-1] != '"' {
			return "", fmt.Errorf("token value %d must end with quote", tokenType)
		}

		value = strings.TrimSuffix(value, `"`)

	// No quotes. noop
	case TokenTypeTranslatorComment, TokenTypeExtractedComment,
		TokenTypeReference, TokenTypeFlag, TokenTypeMsgctxtPreviousContext,
		TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgidPrevUntStr:
	}

	return value, nil
}
