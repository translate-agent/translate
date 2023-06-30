package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// TODO: For now we can only import XLIFF 1.2 files, export is not working correctly yet.

// XLIFF 1.2 specification: https://docs.oasis-open.org/xliff/v1.2/os/xliff-core.html
// XLIFF 1.2 example: https://localizely.com/xliff-file/?tab=xliff-12

type xliff12 struct {
	XMLName xml.Name    `xml:"urn:oasis:names:tc:xliff:document:1.2 xliff"`
	Version string      `xml:"version,attr"`
	File    xliff12File `xml:"file"`
}

type xliff12File struct {
	SourceLanguage language.Tag `xml:"source-language,attr"`
	TargetLanguage language.Tag `xml:"target-language,attr"`
	Body           bodyElement  `xml:"body"`
}

type bodyElement struct {
	TransUnits []transUnit `xml:"trans-unit"`
}

type transUnit struct {
	ID     string `xml:"id,attr"`          // messages.messages[n].ID
	Source string `xml:"source"`           // messages.messages[n].Message (if no target language is set)
	Target string `xml:"target,omitempty"` // messages.messages[n].Message (if target language is set)
	Note   string `xml:"note,omitempty"`   // messages.messages[n].Description
	// No unified standard about storing fuzzy values
}

// FromXliff12 converts serialized data from the XML data in the XLIFF 1.2 format into a model.Messages struct.
func FromXliff12(data []byte) (model.Messages, error) {
	var xlf xliff12
	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal to xliff12: %w", err)
	}

	messages := model.Messages{
		Language: xlf.File.SourceLanguage,
		Messages: make([]model.Message, 0, len(xlf.File.Body.TransUnits)),
	}

	// Check if file has a target language set
	isTranslated := xlf.File.TargetLanguage != language.Und
	if isTranslated {
		messages.Language = xlf.File.TargetLanguage
	}

	getMessage := func(t transUnit) string {
		if isTranslated {
			return t.Target
		}

		return t.Source
	}

	for _, unit := range xlf.File.Body.TransUnits {
		messages.Messages = append(messages.Messages, model.Message{
			ID:          unit.ID,
			Message:     convertToMessageFormatSingular(getMessage(unit)),
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
