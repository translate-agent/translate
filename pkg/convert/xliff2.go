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
	File    xliff2File   `xml:"file"`
}
type xliff2File struct {
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

// FromXliff2 converts serialized data from the XML data in the XLIFF 2 format into a model.Messages struct.
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
			Message:     convertToMessageFormatSingular(unit.Source),
			Description: findDescription(unit),
			Positions:   positionsFromXliff2(*unit.Notes),
		})
	}

	return messages, nil
}

// ToXliff2 converts a model.Messages struct into a byte slice in the XLIFF 2 format.
func ToXliff2(messages model.Messages) ([]byte, error) {
	xlf := xliff2{
		Version: "2.0",
		SrcLang: messages.Language,
		File: xliff2File{
			Units: make([]unit, 0, len(messages.Messages)),
		},
	}

	for _, msg := range messages.Messages {
		notes := positionsToXliff2(msg.Positions)

		if msg.Description != "" {
			if notes == nil {
				notes = &[]note{{Category: "description", Content: msg.Description}}
			} else {
				*notes = append(*notes, note{Category: "description", Content: msg.Description})
			}
		}

		xlf.File.Units = append(xlf.File.Units, unit{
			ID:     msg.ID,
			Source: removeEnclosingBrackets(msg.Message),
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

// helpers

// positionsFromXliff2 extracts line positions from unit []notes.
func positionsFromXliff2(notes []note) model.Positions {
	var positions model.Positions

	for _, note := range notes {
		if note.Category == "location" {
			positions = append(positions, note.Content)
		}
	}

	return positions
}

// positionsFromXliff2 transforms line positions to location []note.
func positionsToXliff2(positions model.Positions) *[]note {
	if len(positions) == 0 {
		return nil
	}

	notes := make([]note, 0, len(positions))

	for _, pos := range positions {
		notes = append(notes, note{Category: "location", Content: pos})
	}

	return &notes
}
