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

// Lex function performs lexical analysis on the input by reading lines from the reader
// and parsing each line using the parseLine function.
func Lex(r io.Reader) ([]Token, error) {
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
		token = TokenTypeHeaderGeneratedBy
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
	// Special case for multiline strings, i.e. msgid, msgstr, msgid_plural, msgstr[*] and headers
	case strings.HasPrefix(line, `"`):
		return nil, parseMultilineValue(line, tokens)
	}

	return parseToken(line, token)
}

// parseToken function parses the value of the token based on the token type.
// On PluralMsgStr token type, it also tries parses the plural index.
func parseToken(line string, tokenType TokenType) (*Token, error) {
	token := Token{Type: tokenType, Value: parseValue(line)}

	// We assume that headers token have a newline at the end so we must trim it.
	if tokenType < TokenTypeMsgCtxt {
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
		return nil, fmt.Errorf("convert '%s' to int: %w", indexStr, err)
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
func parseValue(line string) string {
	n := 2
	fields := strings.SplitN(line, " ", n) // prefix and value, e.g. fields[0] = msgid, fields[1] = "text", etc.

	// No value, only prefix
	if len(fields) != n {
		return ""
	}

	return strings.Trim(fields[1], `"`)
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
		lastToken.Value += "\n" + strings.Trim(line, `"`)

	case TokenTypeMsgCtxt, TokenTypeTranslatorComment, TokenTypeExtractedComment,
		TokenTypeReference, TokenTypeFlag, TokenTypeMsgctxtPreviousContext,
		TokenTypeMsgidPluralPrevUntStrPlural, TokenTypeMsgidPrevUntStr:
		return fmt.Errorf("unsupported multiline string for token type: '%d'", lastToken.Type)
	}

	return nil
}
