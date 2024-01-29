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

// Line prefix constants.
const (
	// Header line prefixes.
	headerLanguagePrefix                = "\"Language:"
	headerPluralFormsPrefix             = "\"Plural-Forms:"
	headerTranslatorPrefix              = "\"Translator:"
	headerProjectIDVersionPrefix        = "\"Project-Id-Version:"
	headerPOTCreationDatePrefix         = "\"POT-Creation-Date:"
	headerPORevisionDatePrefix          = "\"PO-Revision-Date:"
	headerLastTranslatorPrefix          = "\"Last-Translator:"
	headerLanguageTeamPrefix            = "\"Language-Team:"
	headerMIMEVersionPrefix             = "\"MIME-Version:"
	headerContentTypePrefix             = "\"Content-Type:"
	headerContentTransferEncodingPrefix = "\"Content-Transfer-Encoding:"
	headerReportMsgidBugsToPrefix       = "\"Report-Msgid-Bugs-To:"
	headerXGeneratorPrefix              = "\"X-Generator:"
	headerGeneratedByPrefix             = "\"Generated-By:"
	// Message line prefixes.
	msgCtxtPrefix                     = "msgctxt"
	msgIDPrefix                       = "msgid"
	msgStrPrefix                      = "msgstr"
	pluralMsgIDPrefix                 = "msgid_plural"
	pluralMsgStrPrefix                = "msgstr["
	translatorCommentPrefix           = "#"
	extractedCommentPrefix            = "#."
	referenceCommentPrefix            = "#:"
	flagPrefix                        = "#,"
	msgctxtPreviousCtxPrefix          = "#| msgctxt"
	msgidPluralPrevUntStrPluralPrefix = "#| msgid_plural"
	msgidPrevUntStrPrefix             = "#| msgid"
)

var pluralMsgStrRegex = regexp.MustCompile(`^msgstr\[(?:\d+|\*)\]`)

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
func parseLine(line string, tokens []Token) (token *Token, err error) { //nolint:cyclop
	var (
		tokenType  TokenType
		linePrefix string
	)

	switch {
	default:
		return nil, errors.New("unknown line prefix")
	case strings.HasPrefix(line, headerLanguagePrefix):
		tokenType, linePrefix = TokenTypeHeaderLanguage, headerLanguagePrefix
	case strings.HasPrefix(line, headerPluralFormsPrefix):
		tokenType, linePrefix = TokenTypeHeaderPluralForms, headerPluralFormsPrefix
	case strings.HasPrefix(line, headerTranslatorPrefix):
		tokenType, linePrefix = TokenTypeHeaderTranslator, headerTranslatorPrefix
	case strings.HasPrefix(line, headerProjectIDVersionPrefix):
		tokenType, linePrefix = TokenTypeHeaderProjectIDVersion, headerProjectIDVersionPrefix
	case strings.HasPrefix(line, headerPOTCreationDatePrefix):
		tokenType, linePrefix = TokenTypeHeaderPOTCreationDate, headerPOTCreationDatePrefix
	case strings.HasPrefix(line, headerPORevisionDatePrefix):
		tokenType, linePrefix = TokenTypeHeaderPORevisionDate, headerPORevisionDatePrefix
	case strings.HasPrefix(line, headerLastTranslatorPrefix):
		tokenType, linePrefix = TokenTypeHeaderLastTranslator, headerLastTranslatorPrefix
	case strings.HasPrefix(line, headerLanguageTeamPrefix):
		tokenType, linePrefix = TokenTypeHeaderLanguageTeam, headerLanguageTeamPrefix
	case strings.HasPrefix(line, headerMIMEVersionPrefix):
		tokenType, linePrefix = TokenTypeHeaderMIMEVersion, headerMIMEVersionPrefix
	case strings.HasPrefix(line, headerContentTypePrefix):
		tokenType, linePrefix = TokenTypeHeaderContentType, headerContentTypePrefix
	case strings.HasPrefix(line, headerContentTransferEncodingPrefix):
		tokenType, linePrefix = TokenTypeHeaderContentTransferEncoding, headerContentTransferEncodingPrefix
	case strings.HasPrefix(line, headerReportMsgidBugsToPrefix):
		tokenType, linePrefix = TokenTypeHeaderReportMsgidBugsTo, headerReportMsgidBugsToPrefix
	case strings.HasPrefix(line, headerXGeneratorPrefix):
		tokenType, linePrefix = TokenTypeHeaderXGenerator, headerXGeneratorPrefix
	case strings.HasPrefix(line, headerGeneratedByPrefix):
		tokenType, linePrefix = TokenTypeHeaderGeneratedBy, headerGeneratedByPrefix
	case strings.HasPrefix(line, msgCtxtPrefix):
		tokenType, linePrefix = TokenTypeMsgCtxt, msgCtxtPrefix
	case strings.HasPrefix(line, pluralMsgIDPrefix):
		tokenType, linePrefix = TokenTypePluralMsgID, pluralMsgIDPrefix
	case strings.HasPrefix(line, msgIDPrefix):
		tokenType, linePrefix = TokenTypeMsgID, msgIDPrefix
	case strings.HasPrefix(line, pluralMsgStrPrefix):
		if tokenType, linePrefix = TokenTypePluralMsgStr,
			pluralMsgStrRegex.FindString(line); linePrefix == "" {
			return nil, fmt.Errorf("invalid syntax for token type '%d'", tokenType)
		}
	case strings.HasPrefix(line, msgStrPrefix):
		tokenType, linePrefix = TokenTypeMsgStr, msgStrPrefix
	case strings.HasPrefix(line, extractedCommentPrefix):
		tokenType, linePrefix = TokenTypeExtractedComment, extractedCommentPrefix
	case strings.HasPrefix(line, referenceCommentPrefix):
		tokenType, linePrefix = TokenTypeReference, referenceCommentPrefix
	case strings.HasPrefix(line, msgidPluralPrevUntStrPluralPrefix):
		tokenType, linePrefix = TokenTypeMsgidPluralPrevUntStrPlural, msgidPluralPrevUntStrPluralPrefix
	case strings.HasPrefix(line, msgidPrevUntStrPrefix):
		tokenType, linePrefix = TokenTypeMsgidPrevUntStr, msgidPrevUntStrPrefix
	case strings.HasPrefix(line, msgctxtPreviousCtxPrefix):
		tokenType, linePrefix = TokenTypeMsgctxtPreviousContext, msgctxtPreviousCtxPrefix
	case strings.HasPrefix(line, flagPrefix):
		tokenType, linePrefix = TokenTypeFlag, flagPrefix
	case strings.HasPrefix(line, translatorCommentPrefix):
		tokenType, linePrefix = TokenTypeTranslatorComment, translatorCommentPrefix
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
	case TokenTypeMsgID, TokenTypePluralMsgID, TokenTypeMsgStr, TokenTypeMsgCtxt, TokenTypePluralMsgStr,
		TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgctxtPreviousContext, TokenTypeMsgidPrevUntStr:
		if !strings.HasPrefix(value, " ") {
			return "", fmt.Errorf("token type '%d': value must be prefixed with space", tt)
		}

		value = strings.TrimSpace(value)

		// value must be enclosed in double quotes
		if !(strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
			return "", fmt.Errorf("token type '%d': value must be enclosed in double quotation mark - \"\" ", tt)
		}

		value = strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\"")
	case TokenTypeHeaderLanguage, TokenTypeHeaderTranslator, TokenTypeHeaderPluralForms,
		TokenTypeHeaderProjectIDVersion, TokenTypeHeaderPOTCreationDate, TokenTypeHeaderPORevisionDate,
		TokenTypeHeaderLanguageTeam, TokenTypeHeaderLastTranslator, TokenTypeHeaderXGenerator,
		TokenTypeHeaderReportMsgidBugsTo, TokenTypeHeaderMIMEVersion, TokenTypeHeaderContentType,
		TokenTypeHeaderContentTransferEncoding, TokenTypeHeaderGeneratedBy:
		value = strings.TrimSpace(value)

		if !strings.HasSuffix(value, "\"") {
			return "", fmt.Errorf("token type '%d': line must end with double quotation mark - \" ", tt)
		}

		value = strings.TrimSuffix(value, "\"")
	case TokenTypeTranslatorComment, TokenTypeExtractedComment, TokenTypeReference, TokenTypeFlag:
		if !(len(value) == 0 || strings.HasPrefix(value, " ")) {
			return "", fmt.Errorf("token type '%d': value must be prefixed with space", tt)
		}

		value = strings.TrimSpace(value)
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
			return fmt.Errorf("token type '%d': line must end with double quotation mark - \" ", lastToken.Type)
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
