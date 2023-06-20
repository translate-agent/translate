package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// XLIFF 1.2 specification: https://docs.oasis-open.org/xliff/v1.2/os/xliff-core.html
// XLIFF 1.2 example: https://localizely.com/xliff-file/?tab=xliff-12

type xliff12 struct {
	XMLName xml.Name    `xml:"urn:oasis:names:tc:xliff:document:1.2 xliff"`
	Version string      `xml:"version,attr"`
	File    xliff12File `xml:"file"`
}

type xliff12File struct {
	SourceLanguage language.Tag `xml:"source-language,attr"`
	Body           bodyElement  `xml:"body"`
}

type bodyElement struct {
	TransUnits []transUnit `xml:"trans-unit"`
}

type transUnit struct {
	ID     string `xml:"id,attr"`
	Source string `xml:"source"`
	Note   string `xml:"note,omitempty"`
}

// FromXliff12 converts serialized data from the XML data in the XLIFF 1.2 format into a model.Messages struct.
func FromXliff12(data []byte) (model.Messages, error) {
	var xlf xliff12
	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal XLIFF 1.2 formatted XML into xliff12 struct: %w", err)
	}

	messages := model.Messages{
		Language: xlf.File.SourceLanguage,
		Messages: make([]model.Message, 0, len(xlf.File.Body.TransUnits)),
	}

	for _, unit := range xlf.File.Body.TransUnits {
		messages.Messages = append(messages.Messages, model.Message{
			ID:          unit.ID,
			Message:     convertToMessageFormatSingular(unit.Source),
			Description: unit.Note,
		})
	}

	return messages, nil
}

// ToXliff12 converts a model.Messages struct into a byte slice in the XLIFF 1.2 format.
func ToXliff12(messages model.Messages) ([]byte, error) {
	xlf := xliff12{
		Version: "1.2",
		File: xliff12File{
			SourceLanguage: messages.Language,
			Body: bodyElement{
				TransUnits: make([]transUnit, 0, len(messages.Messages)),
			},
		},
	}

	for _, msg := range messages.Messages {
		xlf.File.Body.TransUnits = append(xlf.File.Body.TransUnits, transUnit{
			ID:     msg.ID,
			Source: removeEnclosingBrackets(msg.Message),
			Note:   msg.Description,
		})
	}

	data, err := xml.Marshal(&xlf)
	if err != nil {
		return nil, fmt.Errorf("marshal xliff12 struct to XLIFF 1.2 formatted XML: %w", err)
	}

	dataWithHeader := append([]byte(xml.Header), data...) // prepend generic XML header

	return dataWithHeader, nil
}
