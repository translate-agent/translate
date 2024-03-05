package convert

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"go.expect.digital/mf2"
	ast "go.expect.digital/mf2/parse"
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
	ID            string         `xml:"id,attr"`          // translation.messages[n].ID
	Source        string         `xml:"source"`           // translation.messages[n].Message (if no target language is set)
	Target        string         `xml:"target,omitempty"` // translation.messages[n].Message (if target language is set)
	Note          string         `xml:"note,omitempty"`   // translation.messages[n].Description
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

// FromXliff12 converts serialized data from the XML data in the XLIFF 1.2 format into a model.Translation struct.
func FromXliff12(data []byte, original *bool) (model.Translation, error) {
	var xlf xliff12
	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Translation{}, fmt.Errorf("unmarshal xliff12: %w", err)
	}

	translation := model.Translation{
		Language: xlf.File.TargetLanguage,
		Original: xlf.File.TargetLanguage == language.Und,
		Messages: make([]model.Message, 0, len(xlf.File.Body.TransUnits)),
	}

	// if original is provided override original status in the translation.
	if original != nil {
		translation.Original = *original
	}

	getMessage := func(t transUnit) string { return t.Target }
	status := model.MessageStatusUntranslated

	if translation.Original {
		translation.Language = xlf.File.SourceLanguage
		getMessage = func(t transUnit) string { return t.Source }
		status = model.MessageStatusTranslated
	}

	for _, unit := range xlf.File.Body.TransUnits {
		message, err := mf2.NewBuilder().Text(getMessage(unit)).Build()
		if err != nil {
			return model.Translation{}, fmt.Errorf("convert string to MF2: %w", err)
		}

		translation.Messages = append(translation.Messages, model.Message{
			ID:          unit.ID,
			Message:     message,
			Description: unit.Note,
			Positions:   positionsFromXliff12(unit.ContextGroups),
			Status:      status,
		})
	}

	return translation, nil
}

// ToXliff12 converts a model.Translation struct into a byte slice in the XLIFF 1.2 format.
func ToXliff12(translation model.Translation) ([]byte, error) {
	xlf := xliff12{
		Version: "1.2",
		File: xliff12File{
			Body: bodyElement{
				TransUnits: make([]transUnit, 0, len(translation.Messages)),
			},
		},
	}

	if translation.Original {
		xlf.File.SourceLanguage = translation.Language
	} else {
		xlf.File.TargetLanguage = translation.Language
	}

	for _, msg := range translation.Messages {
		tree, err := ast.Parse(msg.Message)
		if err != nil {
			return nil, fmt.Errorf("parse mf2 message: %w", err)
		}

		switch mf2Msg := tree.Message.(type) {
		case nil:
			msg.Message = ""
		case ast.SimpleMessage:
			msg.Message = patternsToSimpleMsg(mf2Msg)
		case ast.ComplexMessage:
			return nil, errors.New("complex message not supported")
		}

		u := transUnit{
			ID:            msg.ID,
			Note:          msg.Description,
			ContextGroups: positionsToXliff12(msg.Positions),
		}

		if translation.Original {
			u.Source = msg.Message
		} else {
			u.Target = msg.Message
		}

		xlf.File.Body.TransUnits = append(xlf.File.Body.TransUnits, u)
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
