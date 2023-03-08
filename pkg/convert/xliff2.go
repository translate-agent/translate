package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// XLIFF 2 Specification: https://docs.oasis-open.org/xliff/xliff-core/v2.0/os/xliff-core-v2.0-os.html
// XLIFF 2 Example: https://localizely.com/xliff-file/?tab=xliff-20

type xliff2 struct {
	XMLName xml.Name     `xml:"urn:oasis:names:tc:xliff:document:2.0 xliff"`
	Version string       `xml:"version,attr"`
	SrcLang language.Tag `xml:"srcLang,attr"`
	File    file         `xml:"file"`
}
type file struct {
	Units []unit `xml:"unit"`
}

type unit struct {
	ID     string  `xml:"id,attr"`
	Notes  *[]note `xml:"notes>note"` // Set as pointer to avoid empty <notes></notes> when marshalling.
	Source string  `xml:"segment>source"`
}

type note struct {
	Category string `xml:"category,attr"`
	Content  string `xml:",chardata"`
}

func FromXliff2(data []byte) (model.Messages, error) {
	var xlf xliff2
	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal XLIFF 2 formatted XML into xliff2 struct: %w", err)
	}

	messages := model.Messages{Language: xlf.SrcLang, Messages: make([]model.Message, 0, len(xlf.File.Units))}

	findDescription := func(u unit) string {
		for _, note := range *u.Notes {
			if note.Category == "description" {
				return note.Content
			}
		}

		return ""
	}

	for _, unit := range xlf.File.Units {
		messages.Messages = append(messages.Messages, model.Message{
			ID:          unit.ID,
			Message:     unit.Source,
			Description: findDescription(unit),
		})
	}

	return messages, nil
}

func ToXliff2(messages model.Messages) ([]byte, error) {
	xlf := xliff2{
		Version: "2.0",
		SrcLang: messages.Language,
		File: file{
			Units: make([]unit, 0, len(messages.Messages)),
		},
	}

	for _, msg := range messages.Messages {
		var notes *[]note
		if msg.Description != "" {
			notes = &[]note{{Category: "description", Content: msg.Description}}
		}

		xlf.File.Units = append(xlf.File.Units, unit{
			ID:     msg.ID,
			Source: msg.Message,
			Notes:  notes,
		})
	}

	data, err := xml.Marshal(&xlf)
	if err != nil {
		return nil, fmt.Errorf("marshal xliff2 struct to XLIFF 2 formatted XML: %w", err)
	}

	dataWithHeader := append([]byte(xml.Header), data...) // prepend generic XML header

	return dataWithHeader, nil
}
