package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// TODO: For now we can only import XLIFF 2.0 files, export is not working correctly yet.

// XLIFF 2 Specification: https://docs.oasis-open.org/xliff/xliff-core/v2.0/os/xliff-core-v2.0-os.html
// XLIFF 2 Example: https://localizely.com/xliff-file/?tab=xliff-20

type xliff2 struct {
	XMLName xml.Name     `xml:"urn:oasis:names:tc:xliff:document:2.0 xliff"`
	Version string       `xml:"version,attr"`
	SrcLang language.Tag `xml:"srcLang,attr"`
	TrgLang language.Tag `xml:"trgLang,attr"`
	File    xliff2File   `xml:"file"`
}
type xliff2File struct {
	Units []unit `xml:"unit"`
}

type unit struct {
	ID     string  `xml:"id,attr"`                  // messages.messages[n].ID
	Notes  *[]note `xml:"notes>note"`               // Set as pointer to avoid empty <notes></notes> when marshalling.
	Source string  `xml:"segment>source"`           // messages.messages[n].Message (if no target language is set)
	Target string  `xml:"segment>target,omitempty"` // messages.messages[n].Message (if target language is set)
	// No unified standard about storing fuzzy values
}

type note struct {
	Category string `xml:"category,attr"`
	Content  string `xml:",chardata"` // messages.messages[n].Description (if Category == "description")
}

// FromXliff2 converts serialized data from the XML data in the XLIFF 2 format into a model.Messages struct.
// For now original param is ignored.
func FromXliff2(data []byte, original bool) (model.Messages, error) {
	var xlf xliff2

	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal xliff2: %w", err)
	}

	messages := model.Messages{
		Language: xlf.TrgLang,
		Original: xlf.TrgLang == language.Und,
		Messages: make([]model.Message, 0, len(xlf.File.Units)),
	}

	getMessage := func(u unit) string { return u.Target }
	status := model.MessageStatusUntranslated

	if messages.Original {
		messages.Language = xlf.SrcLang
		getMessage = func(u unit) string { return u.Source }
		status = model.MessageStatusTranslated
	}

	findDescription := func(u unit) string {
		for _, note := range *u.Notes {
			if note.Category == "description" {
				return note.Content
			}
		}

		return ""
	}

	for _, unit := range xlf.File.Units {
		message := getMessage(unit)

		messages.Messages = append(messages.Messages, model.Message{
			ID:          unit.ID,
			Message:     convertToMessageFormatSingular(message),
			Description: findDescription(unit),
			Positions:   positionsFromXliff2(unit.Notes),
			Status:      status,
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
		u := unit{
			ID:     msg.ID,
			Source: removeEnclosingBrackets(msg.Message),
			Notes:  positionsToXliff2(msg.Positions),
		}

		if msg.Description != "" {
			if u.Notes == nil {
				u.Notes = &[]note{{Category: "description", Content: msg.Description}}
			} else {
				*u.Notes = append(*u.Notes, note{Category: "description", Content: msg.Description})
			}
		}

		xlf.File.Units = append(xlf.File.Units, u)
	}

	data, err := xml.Marshal(&xlf)
	if err != nil {
		return nil, fmt.Errorf("marshal xliff2: %w", err)
	}

	return append([]byte(xml.Header), data...), nil
}

// helpers

// positionsFromXliff2 extracts line positions from unit []note.
func positionsFromXliff2(notes *[]note) model.Positions {
	if notes == nil {
		return nil
	}

	var positions model.Positions

	for _, note := range *notes {
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
