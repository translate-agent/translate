package convert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"go.expect.digital/mf2"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// TODO: For now we can only import XLIFF 2.0 files, export is not working correctly yet.

// This implementation follows v2.0 specification (last updated 2014-08-05).
// XLIFF 2.0 Specification: https://docs.oasis-open.org/xliff/xliff-core/v2.0/os/xliff-core-v2.0-os.html

// NOTE: Xliff 2.0 has no unified standard for storing fuzzy values, plurals, gender specific text.

// List of elements found in source/target:
// For more details: https://docs.oasis-open.org/xliff/xliff-core/v2.0/os/xliff-core-v2.0-os.html#source
const (
	cp  = "cp"  // Unicode character that is invalid in XML
	ph  = "ph"  // standalone code of the original format
	pc  = "pc"  // well-formed spanning original code
	sc  = "sc"  // start of a spanning original code
	ec  = "ec"  // end of a spanning original code
	mrk = "mrk" // annotation pertaining to the marked span
	sm  = "sm"  // start marker of an annotation
	em  = "em"  // end marker of an annotation
)

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
	ID    string  `xml:"id,attr"`    // translation.messages[n].ID
	Notes *[]note `xml:"notes>note"` // set as pointer to avoid empty <notes></notes> when marshalling
	// NOTE: OriginalData currently is unused.
	OriginalData *[]data `xml:"originalData>data"` // contains the original data for given inline code
	Source       message `xml:"segment>source"`    // translation.messages[n].Message (if no target language is set)
	Target       message `xml:"segment>target"`    // translation.messages[n].Message (if target language is set)
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

// FromXliff2 converts serialized data from the XML data in the XLIFF 2 format into a model.Translation struct.
func FromXliff2(data []byte, original *bool) (model.Translation, error) {
	var xlf xliff2

	if err := xml.Unmarshal(data, &xlf); err != nil {
		return model.Translation{}, fmt.Errorf("unmarshal xliff2: %w", err)
	}

	translation := model.Translation{
		Original: xlf.TrgLang == language.Und,
		Messages: make([]model.Message, 0, len(xlf.File.Units)),
	}

	// if original is provided override original status in the translation.
	if original != nil {
		translation.Original = *original
	}

	if *original {
		translation.Language = xlf.SrcLang
	} else {
		translation.Language = xlf.TrgLang
	}

	for i := range xlf.File.Units {
		msg, err := messageFromUnit(xlf.File.Units[i], translation.Original)
		if err != nil {
			return model.Translation{}, fmt.Errorf("message from unit: %w", err)
		}

		translation.Messages = append(translation.Messages, *msg)
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

	// TODO: implement MF2 to Xliff 2.0 conversion.
	for _, msg := range translation.Messages {
		u, err := messageToUnit(msg, translation.Original)
		if err != nil {
			return nil, fmt.Errorf("message to unit: %w", err)
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

// messageFromUnit converts Xliff2.0 unit element to model.Message.
func messageFromUnit(u unit, original bool) (*model.Message, error) {
	var decoder *xml.Decoder

	m := &model.Message{
		ID:          u.ID,
		Description: descriptionsFromXliff2(u),
		Positions:   positionsFromXliff2(u.Notes),
	}

	if original {
		decoder = xml.NewDecoder(bytes.NewBufferString(u.Source.Content))
		m.Status = model.MessageStatusTranslated
	} else {
		decoder = xml.NewDecoder(bytes.NewBufferString(u.Target.Content))
		m.Status = model.MessageStatusUntranslated
	}

	// retrieve MF2 message from content

	elementCount := make(map[string]int)
	message := mf2.NewBuilder()

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("get token: %w", err)
		}

		if token == nil {
			break
		}

		switch t := token.(type) {
		case xml.CharData: // Text
			message.Text(unescapeXML(string(t)))
		case xml.StartElement:
			switch elName := t.Name.Local; elName {
			default:
				// noop
			case cp, ph, sc, ec, sm, em, pc, mrk:
				elementCount[elName]++

				v := "$" + elName + strconv.Itoa(elementCount[elName])
				message.Expr(mf2.Var(v))

				var attributes string

				for i := range t.Attr {
					attributes += fmt.Sprintf(` %s="%s"`, t.Attr[i].Name.Local, t.Attr[i].Value)
				}

				if elName == pc || elName == mrk { // elements with content
					var content string

					if err = decoder.DecodeElement(&content, &t); err != nil {
						return nil, fmt.Errorf("decode element: %w", err)
					}

					message.Local(v, mf2.Literal("<"+elName+attributes+">"+content+"</"+elName+">"))
				} else { // elements without content
					message.Local(v, mf2.Literal("<"+elName+attributes+"/>"))
				}
			}
		}
	}

	m.Message = message.MustBuild()

	return m, nil
}

func messageToUnit(m model.Message, original bool) (unit, error) { //nolint:unparam
	u := unit{
		ID:    m.ID,
		Notes: positionsToXliff2(m.Positions),
	}

	if m.Description != "" {
		if u.Notes == nil {
			u.Notes = &[]note{{Category: "description", Content: m.Description}}
		} else {
			*u.Notes = append(*u.Notes, note{Category: "description", Content: m.Description})
		}
	}

	// TODO: mf2 to xliff2.0

	if original {
		u.Source.Content = ""
	} else {
		u.Target.Content = ""
	}

	return u, nil
}

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

func descriptionsFromXliff2(u unit) string {
	if u.Notes == nil {
		return ""
	}

	descriptions := make([]string, 0, len(*u.Notes))

	for _, note := range *u.Notes {
		if note.Category == "description" {
			descriptions = append(descriptions, note.Content)
		}
	}

	return strings.Join(descriptions, "\n")
}

// unescapeXML replaces XML escape sequences with corresponding chars in text.
func unescapeXML(s string) string {
	return strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&apos;", "'",
	).Replace(s)
}

// escapeXML replaces special chars in text with escape sequences to be XML compliant.
func escapeXML(s string) string { //nolint:unused
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	).Replace(s)
}
