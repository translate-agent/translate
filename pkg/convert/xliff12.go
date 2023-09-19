package convert

import (
	"encoding/xml"
	"fmt"
	"strings"

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
	ID            string         `xml:"id,attr"`          // messages.messages[n].ID
	Source        string         `xml:"source"`           // messages.messages[n].Message (if no target language is set)
	Target        string         `xml:"target,omitempty"` // messages.messages[n].Message (if target language is set)
	Note          string         `xml:"note,omitempty"`   // messages.messages[n].Description
	ContextGroups []contextGroup `xml:"context-group,omitempty"`
	// No unified standard about storing fuzzy values
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
// For now original param is ignored.
func FromXliff12(data []byte, original bool) (model.Messages, error) {
	var xlf xliff12
	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal xliff12: %w", err)
	}

	messages := model.Messages{
		Language: xlf.File.TargetLanguage,
		Original: xlf.File.TargetLanguage == language.Und,
		Messages: make([]model.Message, 0, len(xlf.File.Body.TransUnits)),
	}

	getMessage := func(t transUnit) string { return t.Target }
	status := model.MessageStatusUntranslated

	// Check if a target language is set
	if messages.Original {
		messages.Language = xlf.File.SourceLanguage
		getMessage = func(t transUnit) string { return t.Source }
		status = model.MessageStatusTranslated
	}

	for _, unit := range xlf.File.Body.TransUnits {
		message := getMessage(unit)

		messages.Messages = append(messages.Messages, model.Message{
			ID:          unit.ID,
			Message:     convertToMessageFormatSingular(message),
			Description: unit.Note,
			Positions:   positionsFromXliff12(unit.ContextGroups),
			Status:      status,
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
		return nil, fmt.Errorf("marshal xliff12: %w", err)
	}

	return append([]byte(xml.Header), data...), nil
}

// helpers

// positionsFromXliff12 extracts line positions from []contextGroup.
func positionsFromXliff12(contextGroups []contextGroup) model.Positions {
	var positions model.Positions

	for _, cg := range contextGroups {
		switch cg.Purpose {
		default:
			continue
		case "location":
			if len(cg.Contexts) == 0 {
				continue
			}

			var sourceFile, lineNumber string

			for _, c := range cg.Contexts {
				switch c.Type {
				case "sourcefile":
					sourceFile = c.Content
				case "linenumber":
					lineNumber = c.Content
				}
			}

			if sourceFile != "" && lineNumber != "" {
				positions = append(positions, sourceFile+":"+lineNumber)
			} else if sourceFile != "" {
				positions = append(positions, sourceFile)
			}
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
