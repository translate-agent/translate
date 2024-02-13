package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/mf2"

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
	ID     string  `xml:"id,attr"`                  // translation.messages[n].ID
	Notes  *[]note `xml:"notes>note"`               // Set as pointer to avoid empty <notes></notes> when marshalling.
	Source string  `xml:"segment>source"`           // translation.messages[n].Message (if no target language is set)
	Target string  `xml:"segment>target,omitempty"` // translation.messages[n].Message (if target language is set)
	// No unified standard about storing fuzzy values
}

type note struct {
	Category string `xml:"category,attr"`
	Content  string `xml:",chardata"` // translation.messages[n].Description (if Category == "description")
}

// FromXliff2 converts serialized data from the XML data in the XLIFF 2 format into a model.Translation struct.
func FromXliff2(data []byte, original *bool) (model.Translation, error) {
	var xlf xliff2

	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Translation{}, fmt.Errorf("unmarshal xliff2: %w", err)
	}

	translation := model.Translation{
		Language: xlf.TrgLang,
		Original: xlf.TrgLang == language.Und,
		Messages: make([]model.Message, 0, len(xlf.File.Units)),
	}

	// if original is provided override original status in the translation.
	if original != nil {
		translation.Original = *original
	}

	getMessage := func(u unit) string { return u.Target }
	status := model.MessageStatusUntranslated

	if translation.Original {
		translation.Language = xlf.SrcLang
		getMessage = func(u unit) string { return u.Source }
		status = model.MessageStatusTranslated
	}

	findDescription := func(u unit) string {
		if u.Notes == nil {
			return ""
		}

		for _, note := range *u.Notes {
			if note.Category == "description" {
				return note.Content
			}
		}

		return ""
	}

	for _, unit := range xlf.File.Units {
		message, err := mf2.NewBuilder().Text(getMessage(unit)).Build()
		if err != nil {
			return model.Translation{}, fmt.Errorf("convert string to MF2: %w", err)
		}

		translation.Messages = append(translation.Messages, model.Message{
			ID:          unit.ID,
			Message:     message,
			Description: findDescription(unit),
			Positions:   positionsFromXliff2(unit.Notes),
			Status:      status,
		})
	}

	return translation, nil
}

// ToXliff2 converts a model.Translation struct into a byte slice in the XLIFF 2 format.
func ToXliff2(translation model.Translation) ([]byte, error) {
	xlf := xliff2{
		Version: "2.0",
		SrcLang: translation.Language,
		File: xliff2File{
			Units: make([]unit, 0, len(translation.Messages)),
		},
	}

	if translation.Original {
		xlf.SrcLang = translation.Language
	} else {
		xlf.TrgLang = translation.Language
	}

	for _, msg := range translation.Messages {
		message := "" // TODO: convert msg.Message from MF2 format.

		u := unit{
			ID:    msg.ID,
			Notes: positionsToXliff2(msg.Positions),
		}

		if translation.Original {
			u.Source = message
		} else {
			u.Target = message
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
