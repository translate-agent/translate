package convert

import (
	"encoding/xml"
	"fmt"
	"strings"

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
	ID            string         `xml:"id,attr"`
	Source        string         `xml:"source"`
	Note          string         `xml:"note,omitempty"`
	ContextGroups []contextGroup `xml:"context-group,omitempty"`
}

type contextGroup struct {
	Purpose  string    `xml:"purpose,attr"`
	Contexts []context `xml:"context,omitempty"`
}

type context struct {
	Type    string `xml:"context-type,attr"`
	Content string `xml:",chardata"`
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
			Positions:   positionsFromXliff12(unit.ContextGroups),
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
			ID:            msg.ID,
			Source:        removeEnclosingBrackets(msg.Message),
			Note:          msg.Description,
			ContextGroups: positionsToXliff12(msg.Positions),
		})
	}

	data, err := xml.Marshal(&xlf)
	if err != nil {
		return nil, fmt.Errorf("marshal xliff12 struct to XLIFF 1.2 formatted XML: %w", err)
	}

	dataWithHeader := append([]byte(xml.Header), data...) // prepend generic XML header

	return dataWithHeader, nil
}

// helpers

// positionsFromXliff12 extracts line positions from []contextGroup.
func positionsFromXliff12(contextGroups []contextGroup) model.Positions {
	var positions model.Positions

	for _, cg := range contextGroups {
		if cg.Purpose == "location" {
			var pos string

			for _, c := range cg.Contexts {
				switch c.Type {
				case "sourcefile":
					if len(pos) > 0 {
						pos += ", " + c.Content
					} else {
						pos += c.Content
					}
				case "linenumber":
					pos += ":" + c.Content
				}
			}

			positions = append(positions, strings.Split(pos, ", ")...)
		}
	}

	return positions
}

// positionsToXliff12 transforms model.Positions to location []contextGroup.
func positionsToXliff12(positions model.Positions) []contextGroup {
	contextGroups := make([]contextGroup, 0, len(positions))

	for _, pos := range positions {
		cg := contextGroup{Purpose: "location"}
		parts := strings.Split(pos, ":")

		switch len(parts) {
		default:
			continue
		case 1:
			cg.Contexts = []context{{Type: "sourcefile", Content: parts[0]}}
		case 2: //nolint:gomnd
			cg.Contexts = []context{
				{Type: "sourcefile", Content: parts[0]},
				{Type: "linenumber", Content: parts[1]},
			}
		}

		contextGroups = append(contextGroups, cg)
	}

	return contextGroups
}
