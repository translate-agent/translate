package convert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"

	mf2 "go.expect.digital/translate/pkg/messageformat"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// TODO: For now we can only import XLIFF 2.0 files, export is not working correctly yet.

// XLIFF was standardized by OASIS in 2002.
// Currently latest version is v2.1 (released on 2018-02-13).

// This implementation follows v2.0 specification (last updated 2014-08-05).
// XLIFF 2.0 Specification: https://docs.oasis-open.org/xliff/xliff-core/v2.0/os/xliff-core-v2.0-os.html
// XLIFF 2.0 Example: https://localizely.com/xliff-file/?tab=xliff-20

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
	ID           string  `xml:"id,attr"`           // translation.messages[n].ID
	Notes        *[]note `xml:"notes>note"`        // set as pointer to avoid empty <notes></notes> when marshalling
	OriginalData *[]data `xml:"originalData>data"` // contains the original data for given inline code
	Source       message `xml:"segment>source"`    // translation.messages[n].Message (if no target language is set)
	Target       message `xml:"segment>target"`    // translation.messages[n].Message (if target language is set)

	// NOTE: Xliff 2.0 has no unified standard for storing fuzzy values, plurals, gender specific text.
}

type data struct {
	ID      string `xml:"id,attr"` // required attribute, currently not enforced
	Content string `xml:",chardata"`
}

type note struct {
	Category string `xml:"category,attr"`
	Content  string `xml:",chardata"` // translation.messages[n].Description (if Category == "description")
}

type message struct {
	Content string `xml:",innerxml"`
}

// Xliff 2.0 placeholder element specification:
// https://docs.oasis-open.org/xliff/xliff-core/v2.0/xliff-core-v2.0.html#ph
type placeholder struct {
	Attributes *[]xml.Attr `xml:",any,attr"` // other attributes, refer to URL above for more details

	ID      string `xml:"id,attr"`      // required attribute, currently not enforced
	DataRef string `xml:"dataRef,attr"` // optional attribute, holds the identifier of the <data> element
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

	getMessage := func(u unit) (string, error) {
		m, err := messageFromContent(u.Target.Content, u.OriginalData)
		if err != nil {
			return "", fmt.Errorf("MF2 message from unit target content: %w", err)
		}

		return m, nil
	}

	status := model.MessageStatusUntranslated

	if translation.Original {
		translation.Language = xlf.SrcLang
		getMessage = func(u unit) (string, error) {
			m, err := messageFromContent(u.Source.Content, u.OriginalData)
			if err != nil {
				return "", fmt.Errorf("MF2 message from unit source content: %w", err)
			}

			return m, nil
		}

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
		message, err := getMessage(unit)
		if err != nil {
			return model.Translation{}, fmt.Errorf("get message: %w", err)
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
		message, err := getMsg(msg.Message)
		if err != nil {
			return nil, fmt.Errorf("get message value: %w", err)
		}

		u := unit{
			ID:    msg.ID,
			Notes: positionsToXliff2(msg.Positions),
		}

		if translation.Original {
			u.Source.Content = message
		} else {
			u.Target.Content = message
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

/*
	messageFromContent extracts 'Message Format v2' compliant message from unit source/target content.

Examples:
input:

	"Entries: <ph id="1" dataRef="d1" canCopy="no" canDelete="no" canOverlap="yes"/>!",
	&[]data{{ID: "d1", Content: "%d"}}

output:

	"Entries: {:Placeholder format=printf type=int value=%d id=1 dataRef=d1 canCopy=no canDelete=no canOverlap=yes}\\!",
	nil
*/
func messageFromContent(content string, originalData *[]data) (string, error) {
	var buf bytes.Buffer

	decoder := xml.NewDecoder(bytes.NewBufferString(content))

	for {
		token, err := decoder.Token()

		if errors.Is(err, io.EOF) {
			return buf.String(), nil
		} else if err != nil {
			return "", fmt.Errorf("get token: %w", err)
		}

		switch t := token.(type) {
		default:
			continue
		case xml.CharData:
			buf.WriteString(mf2.EscapeSpecialChars(string(t)))
		case xml.StartElement:
			// NOTE: currently only placeholder elements are supported.
			// TODO: handle other types of elements.
			if t.Name.Local != "ph" {
				continue
			}

			var ph placeholder

			if err = decoder.DecodeElement(&ph, &t); err != nil {
				return "", fmt.Errorf("decode placeholder: %w", err)
			}

			if err := writePlaceholder(&buf, ph, originalData); err != nil {
				return "", fmt.Errorf("write placeholder: %w", err)
			}
		}
	}
}

// writePlaceholder writes MF2 compliant placeholder expression to bytes.Buffer.
func writePlaceholder(buf *bytes.Buffer, ph placeholder, originalData *[]data) error {
	phExpr := mf2.NodeExpr{
		Function: mf2.NodeFunction{
			Name:    "Placeholder",
			Options: make([]mf2.NodeOption, 0, len(*ph.Attributes)+2), //nolint:gomnd
		},
	}

	// include details about format specifier if referenced in the original data.
	if ph.DataRef != "" && originalData != nil {
		for _, data := range *originalData {
			if ph.DataRef == data.ID {
				pf := mf2.GetPlaceholderFormat(data.Content)

				// If placeholder format is not recognized, default to miscellaneous format.
				if pf == nil {
					pf = &mf2.PlaceholderFormat{
						Re:        regexp.MustCompile(`^`),
						NodeExprF: mf2.CreateNodeExpr("misc"),
					}
				}

				expr := pf.NodeExprF(data.Content, pf.Re.FindStringSubmatchIndex(data.Content))
				phExpr.Function.Options = append(phExpr.Function.Options, expr.Function.Options...)
				// include format specifier
				phExpr.Function.Options = append(phExpr.Function.Options, mf2.NodeOption{Name: "value", Value: data.Content})

				break
			}
		}
	}

	// add placeholder attributes to function options
	if ph.DataRef != "" {
		phExpr.Function.Options = append(phExpr.Function.Options, mf2.NodeOption{Name: "id", Value: ph.ID})
	}

	if ph.DataRef != "" {
		phExpr.Function.Options = append(phExpr.Function.Options, mf2.NodeOption{Name: "dataRef", Value: ph.DataRef})
	}

	if ph.Attributes != nil {
		for _, v := range *ph.Attributes {
			phExpr.Function.Options = append(phExpr.Function.Options,
				mf2.NodeOption{Name: v.Name.Local, Value: v.Value})
		}
	}

	b, err := mf2.AST{phExpr}.MarshalText()
	if err != nil {
		return fmt.Errorf("marshal placeholder text: %w", err)
	}

	// write MF2 compliant placeholder expression to bytes.Buffer.
	buf.Write(b)

	return nil
}
