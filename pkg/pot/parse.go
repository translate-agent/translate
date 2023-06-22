package pot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

type headerNode struct {
	Language    language.Tag
	Translator  string
	PluralForms pluralForm
}

type pluralForm struct {
	Plural   string
	NPlurals int
}

type messageNode struct {
	MsgCtxt               string
	MsgId                 string
	MsgIdPlural           string
	TranslatorComment     []string
	ExtractedComment      []string
	Reference             string
	Flag                  string
	MsgCtxtPrevCtxt       string
	MsgIdPrevUntPluralStr string
	MsgIdPrevUnt          string
	MsgStr                []string
}

type Po struct {
	Header   headerNode
	Messages []messageNode
}

// TokensToPo function takes a slice of Token objects and converts them into a Po object representing
// a PO (Portable Object) file. It returns the generated Po object and an error.
func TokensToPo(tokens []Token) (Po, error) {
	var messages []messageNode

	currentMessage := messageNode{}
	header := headerNode{}

	for i, token := range tokens {
		if token.Value == "" && token.Type == MsgStr {
			prevToken, err := previousToken(tokens, i)
			if err != nil {
				return Po{}, fmt.Errorf("get previous token: %w", err)
			}
			// Skip an empty default msgstr in the header if it exists
			if prevToken.Type == MsgId && prevToken.Value == "" {
				continue
			}
		}

		replacer := strings.NewReplacer("\\n ", "\n", "\\n", "\n")
		token.Value = replacer.Replace(token.Value)

		switch token.Type {
		case HeaderLanguage:
			headerLang, err := language.Parse(token.Value)
			if err != nil {
				return Po{}, fmt.Errorf("invalid language tag: %w", err)
			}

			header.Language = headerLang
		case HeaderTranslator:
			header.Translator = token.Value
		case HeaderPluralForms:
			pf, err := parsePluralForms(token.Value)
			if err != nil {
				return Po{}, fmt.Errorf("invalid plural forms: %w", err)
			}

			header.PluralForms = pf
		case MsgCtxt:
			currentMessage.MsgCtxt = token.Value
		case ExtractedComment:
			currentMessage.ExtractedComment = append(currentMessage.ExtractedComment, token.Value)
		case Reference:
			currentMessage.Reference = token.Value
		case Flag:
			currentMessage.Flag = token.Value
		case TranslatorComment:
			currentMessage.TranslatorComment = append(currentMessage.TranslatorComment, token.Value)
		case MsgctxtPreviousContext:
			currentMessage.MsgCtxtPrevCtxt = token.Value
		case MsgidPluralPrevUntStrPlural:
			currentMessage.MsgIdPrevUntPluralStr = token.Value
		case MsgidPrevUntStr:
			currentMessage.MsgIdPrevUnt = token.Value
		case MsgId:
			currentMessage.MsgId = token.Value
		case PluralMsgId:
			currentMessage.MsgIdPlural = token.Value
		case MsgStr:
			currentMessage.MsgStr = []string{token.Value}
			messages = append(messages, currentMessage)
			currentMessage = messageNode{}
		case PluralMsgStr:
			switch {
			case token.Index == 0:
				currentMessage.MsgStr = []string{token.Value}
			case len(currentMessage.MsgStr) == token.Index:
				currentMessage.MsgStr = append(currentMessage.MsgStr, token.Value)
			case len(currentMessage.MsgStr) < token.Index:
				return Po{}, fmt.Errorf("invalid plural string order: %d", token.Index)
			}

			if header.PluralForms.NPlurals == len(currentMessage.MsgStr) {
				messages = append(messages, currentMessage)
				currentMessage = messageNode{}
			}
			// In our model.Messages currently there are no place to store these headers/metadata about translation file.
		case HeaderReportMsgidBugsTo, HeaderProjectIdVersion, HeaderPOTCreationDate, HeaderPORevisionDate,
			HeaderLanguageTeam, HeaderLastTranslator, HeaderXGenerator, HeaderMIMEVersion, HeaderContentType,
			HeaderContentTransferEncoding:
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
func parsePluralForms(s string) (pluralForm, error) {
	var pf pluralForm

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
