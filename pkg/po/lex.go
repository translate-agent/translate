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
	var (
		lineNumber int
		tokens     []Token
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineNumber++

		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		token, err := parseLine(line, tokens)
		if err != nil {
			return nil, fmt.Errorf(`parse line %d "%s": %w`, lineNumber, line, err)
		}

		if token != nil {
			tokens = append(tokens, *token)
		}
	}

	return tokens, nil
}

// parseLine function processes the line based on different prefixes
// and returns a pointer to a Token object and an error.
//
//nolint:gocyclo, cyclop, gosec
func parseLine(line string, tokens []Token) (token *Token, err error) {
	var (
		tokenType  TokenType
		linePrefix string
	)

	switch {
	default:
		return nil, errors.New("unknown line prefix")
	case strings.HasPrefix(line, "\"Language:"):
		tokenType, linePrefix = TokenTypeHeaderLanguage, "\"Language:"
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		tokenType, linePrefix = TokenTypeHeaderPluralForms, "\"Plural-Forms:"
	case strings.HasPrefix(line, "\"Translator:"):
		tokenType, linePrefix = TokenTypeHeaderTranslator, "\"Translator:"
	case strings.HasPrefix(line, "\"Project-Id-Version:"):
		tokenType, linePrefix = TokenTypeHeaderProjectIDVersion, "\"Project-Id-Version:"
	case strings.HasPrefix(line, "\"POT-Creation-Date:"):
		tokenType, linePrefix = TokenTypeHeaderPOTCreationDate, "\"POT-Creation-Date:"
	case strings.HasPrefix(line, "\"PO-Revision-Date:"):
		tokenType, linePrefix = TokenTypeHeaderPORevisionDate, "\"PO-Revision-Date:"
	case strings.HasPrefix(line, "\"Last-Translator:"):
		tokenType, linePrefix = TokenTypeHeaderLastTranslator, "\"Last-Translator:"
	case strings.HasPrefix(line, "\"Language-Team:"):
		tokenType, linePrefix = TokenTypeHeaderLanguageTeam, "\"Language-Team:"
	case strings.HasPrefix(line, "\"MIME-Version:"):
		tokenType, linePrefix = TokenTypeHeaderMIMEVersion, "\"MIME-Version:"
	case strings.HasPrefix(line, "\"Content-Type:"):
		tokenType, linePrefix = TokenTypeHeaderContentType, "\"Content-Type:"
	case strings.HasPrefix(line, "\"Content-Transfer-Encoding:"):
		tokenType, linePrefix = TokenTypeHeaderContentTransferEncoding, "\"Content-Transfer-Encoding:"
	case strings.HasPrefix(line, "\"Report-Msgid-Bugs-To:"):
		tokenType, linePrefix = TokenTypeHeaderReportMsgidBugsTo, "\"Report-Msgid-Bugs-To:"
	case strings.HasPrefix(line, "\"X-Generator:"):
		tokenType, linePrefix = TokenTypeHeaderXGenerator, "\"X-Generator:"
	case strings.HasPrefix(line, "\"Generated-By:"):
		tokenType, linePrefix = TokenTypeHeaderGeneratedBy, "\"Generated-By:"
	case strings.HasPrefix(line, "msgctxt"):
		tokenType, linePrefix = TokenTypeMsgCtxt, "msgctxt"
	case strings.HasPrefix(line, "msgid_plural"):
		tokenType, linePrefix = TokenTypePluralMsgID, "msgid_plural"
	case strings.HasPrefix(line, "msgid"):
		tokenType, linePrefix = TokenTypeMsgID, "msgid"
	case strings.HasPrefix(line, "msgstr["):
		prefix := regexp.MustCompile(`^msgstr\[\d+\]|\[\*\]`).FindString(line)

		if prefix == "" {
			return nil, fmt.Errorf("invalid syntax for token type '%d'", tokenType)
		}

		tokenType, linePrefix = TokenTypePluralMsgStr, prefix
	case strings.HasPrefix(line, "msgstr"):
		tokenType, linePrefix = TokenTypeMsgStr, "msgstr"
	case strings.HasPrefix(line, "#."):
		tokenType, linePrefix = TokenTypeExtractedComment, "#."
	case strings.HasPrefix(line, "#:"):
		tokenType, linePrefix = TokenTypeReference, "#:"
	case strings.HasPrefix(line, "#| msgid_plural"):
		tokenType, linePrefix = TokenTypeMsgidPluralPrevUntStrPlural, "#| msgid_plural"
	case strings.HasPrefix(line, "#| msgid"):
		tokenType, linePrefix = TokenTypeMsgidPrevUntStr, "#| msgid"
	case strings.HasPrefix(line, "#| msgctxt"):
		tokenType, linePrefix = TokenTypeMsgctxtPreviousContext, "#| msgctxt"
	case strings.HasPrefix(line, "#,"):
		tokenType, linePrefix = TokenTypeFlag, "#,"
	case strings.HasPrefix(line, "#"):
		tokenType, linePrefix = TokenTypeTranslatorComment, "#"
	// Special case for multiline strings, i.e. msgid, msgstr, msgid_plural, msgstr[*] and headers
	case strings.HasPrefix(line, `"`):
		if err = parseMultilineValue(line, tokens); err != nil {
			return nil, fmt.Errorf("parse multiline value: %w", err)
		}

		return token, err
	}

	if token, err = parseToken(linePrefix, line, tokenType); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	return token, nil
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

// parseTokenValue extracts and parses the value of a token.
// e.g.
//
//	line -> value
//	"Language: en-US\n" -> en-US\n
//	msgid "some text" -> some text
//	#, python-format -> python-format
//	#: superset/key_value/exceptions.py:54 -> superset/key_value/exceptions.py:54
//
// Returns the parsed token value and a nil error if successful, or an error if parsing fails.
func parseTokenValue(linePrefix, line string, tokenType TokenType) (string, error) {
	value := strings.TrimPrefix(line, linePrefix)

	switch tt := tokenType; tt {
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypeMsgCtxt, TokenTypePluralMsgStr:
		if !strings.HasPrefix(value, " ") {
			return "", fmt.Errorf("token type '%d': value must be prefixed with space", tt)
		}

		value = strings.TrimSpace(value)

		// value must be enclosed in double quotes
		if !(strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
			return "", fmt.Errorf("token type '%d': value must be enclosed in double quotes", tt)
		}

		value = strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\"")
	case TokenTypeHeaderLanguage, TokenTypeHeaderTranslator, TokenTypeHeaderPluralForms,
		TokenTypeHeaderProjectIDVersion, TokenTypeHeaderPOTCreationDate, TokenTypeHeaderPORevisionDate,
		TokenTypeHeaderLanguageTeam, TokenTypeHeaderLastTranslator, TokenTypeHeaderXGenerator,
		TokenTypeHeaderReportMsgidBugsTo, TokenTypeHeaderMIMEVersion, TokenTypeHeaderContentType,
		TokenTypeHeaderContentTransferEncoding, TokenTypeHeaderGeneratedBy:
		value = strings.TrimSpace(value)

		if !strings.HasSuffix(value, "\"") {
			return "", fmt.Errorf("token type '%d': line must end with double quote", tt)
		}

		value = strings.TrimSuffix(value, "\"")
	case TokenTypeTranslatorComment, TokenTypeExtractedComment, TokenTypeReference, TokenTypeFlag:
		if !(len(value) == 0 || strings.HasPrefix(value, " ")) {
			return "", fmt.Errorf("token type '%d': value must be prefixed with space", tt)
		}

		value = strings.TrimSpace(value)
	case TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgctxtPreviousContext, TokenTypeMsgidPrevUntStr:
		if !(len(value) == 0 || strings.HasPrefix(value, " ")) {
			return "", fmt.Errorf("token type '%d': value must be prefixed with space", tt)
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
		return "", fmt.Errorf("unsupported token type '%d'", tt)
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

		if !strings.HasSuffix(line, "\"") {
			return fmt.Errorf("token type '%d': line must end with double quote", lastToken.Type)
		}

		lastToken.Value += "\n" + strings.TrimPrefix(strings.TrimSuffix(line, "\""), "\"")
	case TokenTypeMsgCtxt, TokenTypeTranslatorComment, TokenTypeExtractedComment,
		TokenTypeReference, TokenTypeFlag, TokenTypeMsgctxtPreviousContext,
		TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgidPrevUntStr:
		return fmt.Errorf("token type '%d': unsupported multiline string", lastToken.Type)
	default:
		return fmt.Errorf("unsupported token type '%d'", lastToken.Type)
	}

	return nil
}
