package po

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
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
	TokenTypeHeaderGeneratedBy
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

// lex function performs lexical analysis on the input by reading lines from the reader
// and parsing each line using the parseLine function.
func lex(r io.Reader) ([]Token, error) {
	var tokens []Token

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		token, err := parseLine(line, tokens)
		if err != nil {
			return nil, fmt.Errorf(`parse line "%s": %w`, line, err)
		}

		if token != nil {
			tokens = append(tokens, *token)
		}
	}

	return tokens, nil
}

// parseLine function processes the line based on different prefixes
// and returns a pointer to a Token object and an error.
func parseLine(line string, tokens []Token) (*Token, error) {
	var (
		token      TokenType
		linePrefix string
	)

	//nolint:gosec
	switch {
	default:
		return nil, fmt.Errorf("unknown prefix in '%s'", line)
	case strings.HasPrefix(line, "\"Language:"):
		token, linePrefix = TokenTypeHeaderLanguage, "\"Language:"
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		token, linePrefix = TokenTypeHeaderPluralForms, "\"Plural-Forms:"
	case strings.HasPrefix(line, "\"Translator:"):
		token, linePrefix = TokenTypeHeaderTranslator, "\"Translator:"
	case strings.HasPrefix(line, "\"Project-Id-Version:"):
		token, linePrefix = TokenTypeHeaderProjectIDVersion, "\"Project-Id-Version:"
	case strings.HasPrefix(line, "\"POT-Creation-Date:"):
		token, linePrefix = TokenTypeHeaderPOTCreationDate, "\"POT-Creation-Date:"
	case strings.HasPrefix(line, "\"PO-Revision-Date:"):
		token, linePrefix = TokenTypeHeaderPORevisionDate, "\"PO-Revision-Date:"
	case strings.HasPrefix(line, "\"Last-Translator:"):
		token, linePrefix = TokenTypeHeaderLastTranslator, "\"Last-Translator:"
	case strings.HasPrefix(line, "\"Language-Team:"):
		token, linePrefix = TokenTypeHeaderLanguageTeam, "\"Language-Team:"
	case strings.HasPrefix(line, "\"MIME-Version:"):
		token, linePrefix = TokenTypeHeaderMIMEVersion, "\"MIME-Version:"
	case strings.HasPrefix(line, "\"Content-Type:"):
		token, linePrefix = TokenTypeHeaderContentType, "\"Content-Type:"
	case strings.HasPrefix(line, "\"Content-Transfer-Encoding:"):
		token, linePrefix = TokenTypeHeaderContentTransferEncoding, "\"Content-Transfer-Encoding:"
	case strings.HasPrefix(line, "\"Report-Msgid-Bugs-To:"):
		token, linePrefix = TokenTypeHeaderReportMsgidBugsTo, "\"Report-Msgid-Bugs-To:"
	case strings.HasPrefix(line, "\"X-Generator:"):
		token, linePrefix = TokenTypeHeaderXGenerator, "\"X-Generator:"
	case strings.HasPrefix(line, "\"Generated-By:"):
		token, linePrefix = TokenTypeHeaderGeneratedBy, "\"Generated-By:"
	case strings.HasPrefix(line, "msgctxt"):
		token, linePrefix = TokenTypeMsgCtxt, "msgctxt"
	case strings.HasPrefix(line, "msgid_plural"):
		token, linePrefix = TokenTypePluralMsgID, "msgid_plural"
	case strings.HasPrefix(line, "msgid"):
		token, linePrefix = TokenTypeMsgID, "msgid"
	case strings.HasPrefix(line, "msgstr["):
		prefix := regexp.MustCompile(`^msgstr\[\d+\]|\[\*\]`).FindString(line)

		if prefix == "" {
			return nil, fmt.Errorf("invalid prefix in '%s'", line)
		}

		token, linePrefix = TokenTypePluralMsgStr, prefix
	case strings.HasPrefix(line, "msgstr"):
		token, linePrefix = TokenTypeMsgStr, "msgstr"
	case strings.HasPrefix(line, "#."):
		token, linePrefix = TokenTypeExtractedComment, "#."
	case strings.HasPrefix(line, "#:"):
		token, linePrefix = TokenTypeReference, "#:"
	case strings.HasPrefix(line, "#| msgid_plural"):
		token, linePrefix = TokenTypeMsgidPluralPrevUntStrPlural, "#| msgid_plural"
	case strings.HasPrefix(line, "#| msgid"):
		token, linePrefix = TokenTypeMsgidPrevUntStr, "#| msgid"
	case strings.HasPrefix(line, "#| msgctxt"):
		token, linePrefix = TokenTypeMsgctxtPreviousContext, "#| msgctxt"
	case strings.HasPrefix(line, "#,"):
		token, linePrefix = TokenTypeFlag, "#,"
	case strings.HasPrefix(line, "#"):
		token, linePrefix = TokenTypeTranslatorComment, "#"
	// Special case for multiline strings, i.e. msgid, msgstr, msgid_plural, msgstr[*] and headers
	case strings.HasPrefix(line, `"`):
		return nil, parseMultilineValue(line, tokens)
	}

	return parseToken(linePrefix, line, token)
}

// parseToken function parses the value of the token based on the token type.
// On PluralMsgStr token type, it also tries parses the plural index.
func parseToken(linePrefix, line string, tokenType TokenType) (token *Token, err error) {
	token = &Token{Type: tokenType}

	if token.Value, err = parseTokenValue(linePrefix, line, tokenType); err != nil {
		return nil, fmt.Errorf("parse token value: %w", err)
	}

	// We assume that headers token have a newline at the end so we must trim it.
	if tokenType < TokenTypeMsgCtxt {
		token.Value = strings.TrimSuffix(token.Value, "\\n")
	}

	if tokenType != TokenTypePluralMsgStr {
		return token, nil
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
		return nil, fmt.Errorf("convert '%s' to int: %w", indexStr, err)
	}

	token.Index = index

	return token, nil
}

// parseTokenValue parses the value of the token.
func parseTokenValue(linePrefix, line string, tokenType TokenType) (string, error) {
	value := strings.TrimPrefix(line, linePrefix)

	switch tt := tokenType; tt {
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypeMsgCtxt, TokenTypePluralMsgStr:
		if !strings.HasPrefix(value, " ") {
			return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
		}

		value = strings.TrimSpace(value)

		// value must be enclosed in double quotes
		if !(strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
			return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
		}

		value = strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\"")
	case TokenTypeHeaderLanguage, TokenTypeHeaderTranslator, TokenTypeHeaderPluralForms,
		TokenTypeHeaderProjectIDVersion, TokenTypeHeaderPOTCreationDate, TokenTypeHeaderPORevisionDate,
		TokenTypeHeaderLanguageTeam, TokenTypeHeaderLastTranslator, TokenTypeHeaderXGenerator,
		TokenTypeHeaderReportMsgidBugsTo, TokenTypeHeaderMIMEVersion, TokenTypeHeaderContentType,
		TokenTypeHeaderContentTransferEncoding, TokenTypeHeaderGeneratedBy:
		value = strings.TrimSpace(value)

		if !strings.HasSuffix(value, "\"") {
			return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
		}

		value = strings.TrimSuffix(value, "\"")
	case TokenTypeTranslatorComment, TokenTypeExtractedComment, TokenTypeReference, TokenTypeFlag:
		if !(len(value) == 0 || strings.HasPrefix(value, " ")) {
			return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
		}

		value = strings.TrimSpace(value)
	case TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgctxtPreviousContext, TokenTypeMsgidPrevUntStr:
		if !(len(value) == 0 || strings.HasPrefix(value, " ")) {
			return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
		}

		value = strings.TrimSpace(value)

		if len(value) != 0 {
			value = " " + value
		}

		switch tt { //nolint:exhaustive
		case TokenTypeMsgidPluralPrevUntStrPlural:
			value = "msgid_plural" + value
		case TokenTypeMsgctxtPreviousContext:
			value = "msgctxt" + value
		case TokenTypeMsgidPrevUntStr:
			value = "msgid" + value
		}
	default:
		return "", fmt.Errorf("unsupported token type: '%d'", tokenType)
	}

	return value, nil
}

// parseMultilineValue parses the value of the multiline msgid, msgstr, msgid_plural, msgstr[*]
// and headers tokens.
// Instead of returning a parsed value, it appends it to the last token in the tokens slice.
func parseMultilineValue(line string, tokens []Token) error {
	lastToken := &tokens[len(tokens)-1]

	switch lastToken.Type {
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypePluralMsgStr,
		TokenTypeHeaderLanguage, TokenTypeHeaderTranslator, TokenTypeHeaderPluralForms,
		TokenTypeHeaderProjectIDVersion, TokenTypeHeaderPOTCreationDate, TokenTypeHeaderPORevisionDate,
		TokenTypeHeaderLanguageTeam, TokenTypeHeaderLastTranslator, TokenTypeHeaderXGenerator,
		TokenTypeHeaderReportMsgidBugsTo, TokenTypeHeaderMIMEVersion, TokenTypeHeaderContentType,
		TokenTypeHeaderContentTransferEncoding, TokenTypeHeaderGeneratedBy:
		line = strings.TrimSpace(line)

		// value must be enclosed in double quotes
		if !(strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"")) {
			return fmt.Errorf("invalid syntax in multiline string for token type: '%d'", lastToken.Type)
		}

		lastToken.Value += "\n" + strings.TrimPrefix(strings.TrimSuffix(line, "\""), "\"")
	case TokenTypeMsgCtxt, TokenTypeTranslatorComment, TokenTypeExtractedComment,
		TokenTypeReference, TokenTypeFlag, TokenTypeMsgctxtPreviousContext,
		TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgidPrevUntStr:
		return fmt.Errorf("unsupported multiline string for token type: '%d'", lastToken.Type)
	default:
		return fmt.Errorf("unsupported token type: '%d'", lastToken.Type)
	}

	return nil
}
