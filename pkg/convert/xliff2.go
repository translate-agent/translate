package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type xliff2 struct {
	XMLName xml.Name     `xml:"urn:oasis:names:tc:xliff:document:2.0 xliff"`
	SrcLang language.Tag `xml:"srcLang,attr"`
	File    file         `xml:"file"`
}
type file struct {
	Units []unit `xml:"unit"`
}

type unit struct {
	ID     string `xml:"id,attr"`
	Source string `xml:"segment>source"`
	Notes  []note `xml:"notes>note"`
}

type note struct {
	Category string `xml:"category,attr"`
	Content  string `xml:",chardata"`
}

func FromXliff2(data []byte) (model.Messages, error) {
	var xlf xliff2
	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal XLIFF 2 formatted XML to model.Messages: %w", err)
	}

	messages := model.Messages{Language: xlf.SrcLang, Messages: make([]model.Message, 0, len(xlf.File.Units))}

	findDescription := func(u unit) string {
		for _, note := range u.Notes {
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
