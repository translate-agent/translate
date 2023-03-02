package convert

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// extract-i18n guide and specification: https://angular.io/guide/i18n-common-translation-files
// extract-i18n JSON example: https://github.com/angular/angular/issues/45465
// extract-i18n XLF example: https://angular.io/guide/i18n-common-translation-files#translate-each-translation-file

// ----------------------JSON--------------------------

// Extracted messages with json format contains only id:message.
type ngJSON struct {
	Language     language.Tag      `json:"locale"`
	Translations map[string]string `json:"translations"`
}

func (n ngJSON) fromNgJSON() model.Messages {
	messages := model.Messages{
		Language: n.Language,
		Messages: make([]model.Message, 0, len(n.Translations)),
	}

	for k, v := range n.Translations {
		msg := model.Message{ID: k, Message: v}
		messages.Messages = append(messages.Messages, msg)
	}

	return messages
}

// --------XLF 1.2 (Default for ng extract-i18n)---------

type ngXLF12 struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:xliff:document:1.2 xliff"`
	File    file     `xml:"file"`
}

type file struct {
	SourceLanguage language.Tag `xml:"source-language,attr"`
	Body           bodyElement  `xml:"body"`
}

type bodyElement struct {
	TransUnits []transUnit `xml:"trans-unit"`
}

type transUnit struct {
	ID     string `xml:"id,attr"`
	Source string `xml:"source"`
	Note   string `xml:"note"`
}

func (n ngXLF12) fromNgXLF12() model.Messages {
	messages := model.Messages{
		Language: n.File.SourceLanguage,
		Messages: make([]model.Message, 0, len(n.File.Body.TransUnits)),
	}

	for _, unit := range n.File.Body.TransUnits {
		msg := model.Message{ID: unit.ID, Message: unit.Source, Description: unit.Note}
		messages.Messages = append(messages.Messages, msg)
	}

	return messages
}

// FromNG converts serialized data from the ng extract-i18n tool into a model.Messages struct.
//
// This function supports two input formats:
//   - JSON (produced by running "ng extract-i18n --format json")
//   - XLF (based on the XLIFF 1.2 standard, which is the default format for ng extract-i18n).
func FromNG(data []byte) (model.Messages, error) {
	var from func() model.Messages

	switch firstByte := data[0]; firstByte {
	case '{': // JSON format
		var ngData ngJSON
		if err := json.Unmarshal(data, &ngData); err != nil {
			return model.Messages{}, fmt.Errorf("unmarshal json data: %w", err)
		}

		from = ngData.fromNgJSON

	case '<': // XML format
		var ngData ngXLF12
		if err := xml.Unmarshal(data, &ngData); err != nil {
			return model.Messages{}, fmt.Errorf("unmarshal xml data: %w", err)
		}

		from = ngData.fromNgXLF12

	default:
		return model.Messages{}, errors.New("unsupported file format")
	}

	return from(), nil
}
