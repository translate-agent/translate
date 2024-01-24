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
		token TokenType
		value string
	)

	switch {
	default:
		return nil, fmt.Errorf("unknown prefix in '%s'", line)
	case strings.HasPrefix(line, "\"Language:"):
		token = TokenTypeHeaderLanguage
		value = strings.TrimPrefix(line, "\"Language:")
	case strings.HasPrefix(line, "\"Plural-Forms:"):
		token = TokenTypeHeaderPluralForms
		value = strings.TrimPrefix(line, "\"Plural-Forms:")
	case strings.HasPrefix(line, "\"Translator:"):
		token = TokenTypeHeaderTranslator
		value = strings.TrimPrefix(line, "\"Translator:")
	case strings.HasPrefix(line, "\"Project-Id-Version:"):
		token = TokenTypeHeaderProjectIDVersion
		value = strings.TrimPrefix(line, "\"Project-Id-Version:")
	case strings.HasPrefix(line, "\"POT-Creation-Date"):
		token = TokenTypeHeaderPOTCreationDate
		value = strings.TrimPrefix(line, "\"POT-Creation-Date:")
	case strings.HasPrefix(line, "\"PO-Revision-Date:"):
		token = TokenTypeHeaderPORevisionDate
		value = strings.TrimPrefix(line, "\"PO-Revision-Date:")
	case strings.HasPrefix(line, "\"Last-Translator:"):
		token = TokenTypeHeaderLastTranslator
		value = strings.TrimPrefix(line, "\"Last-Translator:")
	case strings.HasPrefix(line, "\"Language-Team:"):
		token = TokenTypeHeaderLanguageTeam
		value = strings.TrimPrefix(line, "\"Language-Team:")
	case strings.HasPrefix(line, "\"MIME-Version:"):
		token = TokenTypeHeaderMIMEVersion
		value = strings.TrimPrefix(line, "\"MIME-Version:")
	case strings.HasPrefix(line, "\"Content-Type:"):
		token = TokenTypeHeaderContentType
		value = strings.TrimPrefix(line, "\"Content-Type:")
	case strings.HasPrefix(line, "\"Content-Transfer-Encoding:"):
		token = TokenTypeHeaderContentTransferEncoding
		value = strings.TrimPrefix(line, "\"Content-Transfer-Encoding:")
	case strings.HasPrefix(line, "\"Report-Msgid-Bugs-To:"):
		token = TokenTypeHeaderReportMsgidBugsTo
		value = strings.TrimPrefix(line, "\"Report-Msgid-Bugs-To:")
	case strings.HasPrefix(line, "\"X-Generator:"):
		token = TokenTypeHeaderXGenerator
		value = strings.TrimPrefix(line, "\"X-Generator:")
	case strings.HasPrefix(line, "\"Generated-By:"):
		token = TokenTypeHeaderGeneratedBy
		value = strings.TrimPrefix(line, "\"Generated-By:")
	case strings.HasPrefix(line, "msgctxt"):
		token = TokenTypeMsgCtxt
		value = strings.TrimPrefix(line, "msgctxt")
	case strings.HasPrefix(line, "msgid_plural"):
		token = TokenTypePluralMsgID
		value = strings.TrimPrefix(line, "msgid_plural")
	case strings.HasPrefix(line, "msgid"):
		token = TokenTypeMsgID
		value = strings.TrimPrefix(line, "msgid")
	case strings.HasPrefix(line, "msgstr["):
		token = TokenTypePluralMsgStr
		value = strings.TrimPrefix(line, "msgstr")
	case strings.HasPrefix(line, "msgstr"):
		token = TokenTypeMsgStr
		value = strings.TrimPrefix(line, "msgstr")
	case strings.HasPrefix(line, "#."):
		token = TokenTypeExtractedComment
		value = strings.TrimPrefix(line, "#.")
	case strings.HasPrefix(line, "#:"):
		token = TokenTypeReference
		value = strings.TrimPrefix(line, "#:")
	case strings.HasPrefix(line, "#| msgid_plural"):
		token = TokenTypeMsgidPluralPrevUntStrPlural
		value = strings.TrimPrefix(line, "#|")
	case strings.HasPrefix(line, "#| msgid"):
		token = TokenTypeMsgidPrevUntStr
		value = strings.TrimPrefix(line, "#|")
	case strings.HasPrefix(line, "#| msgctxt"):
		token = TokenTypeMsgctxtPreviousContext
		value = strings.TrimPrefix(line, "#|")
	case strings.HasPrefix(line, "#,"):
		token = TokenTypeFlag
		value = strings.TrimPrefix(line, "#,")
	case strings.HasPrefix(line, "#"):
		token = TokenTypeTranslatorComment
		value = strings.TrimPrefix(line, "#")
	// Special case for multiline strings, i.e. msgid, msgstr, msgid_plural, msgstr[*] and headers
	case strings.HasPrefix(line, `"`):
		return nil, parseMultilineValue(line, tokens)
	}

	return parseToken(line, value, token)
}

// parseToken function parses the value of the token based on the token type.
// On PluralMsgStr token type, it also tries parses the plural index.
func parseToken(line, value string, tokenType TokenType) (token *Token, err error) {
	token = &Token{Type: tokenType}

	if token.Value, err = parseTokenValue(tokenType, value); err != nil {
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
func parseTokenValue(tokenType TokenType, value string) (string, error) {
	switch tokenType {
	case TokenTypePluralMsgStr:
		re := regexp.MustCompile(`^\[\d+\]\s+|\[\*\]\s+`)

		if re.MatchString(value) {
			value = strings.TrimPrefix(value, re.FindString(value))
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			return strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\""), nil
		}
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypeMsgCtxt:
		if strings.HasPrefix(value, " ") {
			value = strings.TrimSpace(value)

			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				return strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\""), nil
			}
		}
	case TokenTypeHeaderLanguage, TokenTypeHeaderTranslator, TokenTypeHeaderPluralForms,
		TokenTypeHeaderProjectIDVersion, TokenTypeHeaderPOTCreationDate, TokenTypeHeaderPORevisionDate,
		TokenTypeHeaderLanguageTeam, TokenTypeHeaderLastTranslator, TokenTypeHeaderXGenerator,
		TokenTypeHeaderReportMsgidBugsTo, TokenTypeHeaderMIMEVersion, TokenTypeHeaderContentType,
		TokenTypeHeaderContentTransferEncoding, TokenTypeHeaderGeneratedBy:
		value = strings.TrimSpace(value)

		if strings.HasSuffix(value, "\"") {
			return strings.TrimSuffix(value, "\""), nil
		}

		return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
	case TokenTypeTranslatorComment,
		TokenTypeExtractedComment, TokenTypeReference, TokenTypeFlag,
		TokenTypeMsgctxtPreviousContext, TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgidPrevUntStr:
		if len(value) == 0 || strings.HasPrefix(value, " ") {
			return strings.TrimSpace(value), nil
		}
	default:
		return "", fmt.Errorf("unsupported token type: '%d'", tokenType)
	}

	return "", fmt.Errorf("invalid syntax in line string for token type: '%d'", tokenType)
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

		if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
			lastToken.Value += "\n" + strings.TrimPrefix(strings.TrimSuffix(line, "\""), "\"")
		} else {
			return fmt.Errorf("invalid syntax in multiline string for token type: '%d'", lastToken.Type)
		}
	case TokenTypeMsgCtxt, TokenTypeTranslatorComment, TokenTypeExtractedComment,
		TokenTypeReference, TokenTypeFlag, TokenTypeMsgctxtPreviousContext,
		TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgidPrevUntStr:
		return fmt.Errorf("unsupported multiline string for token type: '%d'", lastToken.Type)
	}

	return nil
}
