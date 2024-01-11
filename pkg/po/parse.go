package po

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

type HeaderNode struct {
	Language    language.Tag
	Translator  string
	PluralForms PluralForm
}

func (h *HeaderNode) marshal() []byte {
	var sb bytes.Buffer

	sb.WriteString("msgid \"\"\nmsgstr \"\"\n")

	if h.Language != language.Und {
		sb.WriteString(fmt.Sprintf("\"Language: %s\\n\"\n", h.Language.String()))
	}

	if h.Translator != "" {
		sb.WriteString(fmt.Sprintf("\"Last-Translator: %s\\n\"\n", h.Translator))
	}

	if pluralForms := h.PluralForms.marshal(); len(pluralForms) > 0 {
		sb.Write(pluralForms)
		sb.WriteByte('\n')
	}

	return sb.Bytes()
}

type PluralForm struct {
	Plural   string
	NPlurals int
}

func (pf *PluralForm) marshal() []byte {
	if pf.Plural == "" && pf.NPlurals == 0 {
		return []byte{}
	}

	return []byte(fmt.Sprintf("\"Plural-Forms: nplurals=%d; %s\\n\"", pf.NPlurals, pf.Plural))
}

type MessageNode struct {
	MsgCtxt               string
	MsgID                 string
	MsgIDPlural           string
	TranslatorComment     []string
	ExtractedComment      []string
	References            []string
	Flags                 []string
	MsgCtxtPrevCtxt       string
	MsgIDPrevUntPluralStr string
	MsgIDPrevUnt          string
	MsgStr                []string
}

func (m *MessageNode) marshal() []byte {
	var sb bytes.Buffer

	// quoteLines function splits a string into multiple lines and wraps each line in double quotes.
	quoteLines := func(s string) string {
		split := strings.Split(s, "\n")
		for i, line := range split {
			split[i] = fmt.Sprintf("\"%s\"", line)
		}

		return strings.Join(split, "\n")
	}

	sb.WriteByte('\n') // empty line before each message

	for _, reference := range m.References {
		sb.WriteString(fmt.Sprintf("#: %s\n", reference))
	}

	for _, ec := range m.ExtractedComment {
		sb.WriteString(fmt.Sprintf("#. %s\n", ec))
	}

	for _, flag := range m.Flags {
		sb.WriteString(fmt.Sprintf("#, %s\n", flag))
	}

	if m.MsgID != "" {
		sb.WriteString(fmt.Sprintf("msgid %s\n", quoteLines(m.MsgID)))
	}

	if m.MsgIDPlural != "" {
		sb.WriteString(fmt.Sprintf("msgid_plural %s\n", quoteLines(m.MsgIDPlural)))
	}

	switch len(m.MsgStr) {
	case 0:
		sb.WriteString("msgstr \"\"\n")
	case 1:
		sb.WriteString(fmt.Sprintf("msgstr %s\n", quoteLines(m.MsgStr[0])))
	default:
		for i, ms := range m.MsgStr {
			sb.WriteString(fmt.Sprintf("msgstr[%d] %s\n", i, quoteLines(ms)))
		}
	}

	// TODO: Add support for other fields

	return sb.Bytes()
}

type Po struct {
	Header   HeaderNode
	Messages []MessageNode
}

// Marshal serializes the Po object into a byte slice.
func (p *Po) Marshal() []byte {
	var sb bytes.Buffer

	sb.Write(p.Header.marshal())

	for _, msg := range p.Messages {
		sb.Write(msg.marshal())
	}

	return sb.Bytes()
}

// Parse function takes an io.Reader object and parses the contents into a Po struct
// which represents object representing a Portable Object file.
func Parse(r io.Reader) (Po, error) {
	tokens, err := lex(r)
	if err != nil {
		return Po{}, fmt.Errorf("lex: %w", err)
	}

	po, err := tokensToPo(tokens)
	if err != nil {
		return Po{}, fmt.Errorf("tokens to po: %w", err)
	}

	return po, nil
}

// tokensToPo function takes a slice of Token objects and converts them into a Po object representing
// a PO (Portable Object) file. It returns the generated Po object and an error.
func tokensToPo(tokens []Token) (Po, error) {
	var messages []MessageNode

	currentMessage := MessageNode{}

	var header HeaderNode

	for i, token := range tokens {
		if token.Value == "" && token.Type == TokenTypeMsgStr {
			prevToken, err := previousToken(tokens, i)
			if err != nil {
				return Po{}, fmt.Errorf("get previous token: %w", err)
			}
			// Skip an empty default msgstr in the header if it exists
			if prevToken.Type == TokenTypeMsgID && prevToken.Value == "" {
				continue
			}
		}

		switch token.Type {
		case TokenTypeHeaderLanguage:
			headerLang, err := language.Parse(token.Value)
			if err != nil {
				return Po{}, fmt.Errorf("invalid language: %w", err)
			}

			header.Language = headerLang
		case TokenTypeHeaderTranslator:
			header.Translator = token.Value
		case TokenTypeHeaderPluralForms:
			pf, err := parsePluralForms(token.Value)
			if err != nil {
				return Po{}, fmt.Errorf("invalid plural forms: %w", err)
			}

			header.PluralForms = pf
		case TokenTypeMsgCtxt:
			currentMessage.MsgCtxt = token.Value
		case TokenTypeExtractedComment:
			currentMessage.ExtractedComment = append(currentMessage.ExtractedComment, token.Value)
		case TokenTypeReference:
			currentMessage.References = append(currentMessage.References, token.Value)
		case TokenTypeFlag:
			currentMessage.Flags = append(currentMessage.Flags, token.Value)
		case TokenTypeTranslatorComment:
			currentMessage.TranslatorComment = append(currentMessage.TranslatorComment, token.Value)
		case TokenTypeMsgctxtPreviousContext:
			currentMessage.MsgCtxtPrevCtxt = token.Value
		case TokenTypeMsgidPluralPrevUntStrPlural:
			currentMessage.MsgIDPrevUntPluralStr = token.Value
		case TokenTypeMsgidPrevUntStr:
			currentMessage.MsgIDPrevUnt = token.Value
		case TokenTypeMsgID:
			currentMessage.MsgID = token.Value
		case TokenTypePluralMsgID:
			currentMessage.MsgIDPlural = token.Value
		case TokenTypeMsgStr:
			currentMessage.MsgStr = []string{token.Value}
			messages = append(messages, currentMessage)
			currentMessage = MessageNode{}
		case TokenTypePluralMsgStr:
			currentMessage.MsgStr = append(currentMessage.MsgStr, token.Value)
			// if next token is not plural msgstr, then we have all the plural msgstrs
			if i+1 >= len(tokens) || tokens[i+1].Type != TokenTypePluralMsgStr {
				messages = append(messages, currentMessage)
				currentMessage = MessageNode{}
			}
			// In our model.Translation currently there are no place to store these headers/metadata about translation file.
		case TokenTypeHeaderReportMsgidBugsTo,
			TokenTypeHeaderProjectIDVersion,
			TokenTypeHeaderPOTCreationDate,
			TokenTypeHeaderPORevisionDate,
			TokenTypeHeaderLanguageTeam,
			TokenTypeHeaderLastTranslator,
			TokenTypeHeaderXGenerator,
			TokenTypeHeaderMIMEVersion,
			TokenTypeHeaderContentType,
			TokenTypeHeaderContentTransferEncoding,
			TokenTypeHeaderGeneratedBy:
			continue
		}
	}

	if len(messages) == 0 {
		return Po{}, errors.New("invalid po file: no messages found")
	}

	return Po{
		Header:   header,
		Messages: messages,
	}, nil
}

// parsePluralForms function splits the input string into two parts using the separator "; ".
// The first part represents the "nplurals" information and is further split using "=" as the separator.
// The second part represents the plural expression and is trimmed of leading and trailing whitespace.
// The function converts the parsed "nplurals" value to an integer and assigns it to the pluralForm object.
func parsePluralForms(s string) (PluralForm, error) {
	var pf PluralForm

	var err error

	pfArgCount := 2

	parts := strings.Split(s, "; ")
	if len(parts) != pfArgCount {
		return pf, errors.New("invalid plural forms format")
	}

	nPluralsParts := strings.Split(strings.TrimSpace(parts[0]), "=")
	if len(nPluralsParts) != pfArgCount {
		return pf, errors.New("invalid nplurals part")
	}

	pf.NPlurals, err = strconv.Atoi(nPluralsParts[1])
	if err != nil {
		return pf, fmt.Errorf("invalid nplurals value: %w", err)
	}

	pf.Plural = strings.TrimSpace(parts[1])

	return pf, nil
}

// previousToken function takes a slice of Token objects and an index representing the current position in the slice.
// It returns the previous Token object relative to the given index.
func previousToken(tokens []Token, index int) (Token, error) {
	if index == 0 {
		return Token{}, errors.New("no previous token")
	}

	return tokens[index-1], nil
}
