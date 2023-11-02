package convert

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.expect.digital/translate/pkg/messageformat"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/pot"
)

type poTag string

const (
	MsgID        poTag = "msgid"
	PluralID     poTag = "msgid_plural"
	MsgStrPlural poTag = "msgstr[%d]"
	MsgStr       poTag = "msgstr"
)

// ToPot function takes a model.Translation structure,
// writes the language information and each message to a buffer in the POT file format,
// and returns the buffer contents as a byte slice representing the POT file.
func ToPot(t model.Translation) ([]byte, error) {
	var b bytes.Buffer

	if _, err := fmt.Fprintf(&b, "msgid \"\"\nmsgstr \"\"\n\"Language: %s\\n\"\n", t.Language); err != nil {
		return nil, fmt.Errorf("write language: %w", err)
	}

	for i, message := range t.Messages {
		if err := writeMessage(&b, i, message); err != nil {
			return nil, fmt.Errorf("write message: %w", err)
		}
	}

	return b.Bytes(), nil
}

// FromPot function parses a POT file by tokenizing and converting it into a pot.Po structure.
func FromPot(b []byte, original bool) (model.Translation, error) {
	const pluralCountLimit = 2

	tokens, err := pot.Lex(bytes.NewReader(b))
	if err != nil {
		return model.Translation{}, fmt.Errorf("divide po file to tokens: %w", err)
	}

	po, err := pot.TokensToPo(tokens)
	if err != nil {
		return model.Translation{}, fmt.Errorf("convert tokens to pot.Po: %w", err)
	}

	if po.Header.PluralForms.NPlurals > pluralCountLimit {
		return model.Translation{}, errors.New("plural forms with more than 2 forms are not implemented yet")
	}

	messages := make([]model.Message, 0, len(po.Messages))

	singularValue := func(v pot.MessageNode) string { return v.MsgStr[0] }
	pluralValue := func(v pot.MessageNode) []string { return v.MsgStr }
	getStatus := func(v pot.MessageNode) model.MessageStatus {
		if strings.Contains(v.Flag, "fuzzy") {
			return model.MessageStatusFuzzy
		}

		return model.MessageStatusUntranslated
	}

	if original {
		singularValue = func(v pot.MessageNode) string { return v.MsgID }
		pluralValue = func(v pot.MessageNode) []string { return []string{v.MsgID, v.MsgIDPlural} }
		getStatus = func(_ pot.MessageNode) model.MessageStatus { return model.MessageStatusTranslated }
	}

	convert := func(v pot.MessageNode) string { return convertToMessageFormatSingular(singularValue(v)) }

	if po.Header.PluralForms.NPlurals == pluralCountLimit {
		convert = func(v pot.MessageNode) string {
			if v.MsgIDPlural == "" || allEmpty(pluralValue(v)) {
				return convertToMessageFormatSingular(singularValue(v))
			}

			return convertPluralsToMessageString(pluralValue(v))
		}
	}

	for _, node := range po.Messages {
		message := model.Message{
			ID:          node.MsgID,
			PluralID:    node.MsgIDPlural,
			Description: strings.Join(node.ExtractedComment, "\n "),
			Positions:   node.References,
			Message:     convert(node),
			Status:      getStatus(node),
		}

		messages = append(messages, message)
	}

	return model.Translation{
		Language: po.Header.Language,
		Messages: messages,
		Original: original,
	}, nil
}

// allEmpty checks if string values in string slice are empty.
func allEmpty(strings []string) bool {
	for _, str := range strings {
		if str != "" {
			return false
		}
	}

	return true
}

// writeToPoTag function selects the appropriate write function (writeDefault or writePlural)
// based on the tag type (poTag) and uses that function to write the tag value to a bytes.Buffer.
func writeToPoTag(b *bytes.Buffer, tag poTag, str string) error {
	write := writeDefault
	if tag == MsgStrPlural {
		write = writePlural
	}

	if err := write(b, tag, str); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// writeDefault parses a default message string into nodes, accumulates the text content,
// handles cases where the default message doesn't have any text, splits the accumulated text into lines,
// and writes the lines as a multiline tag value to a bytes.Buffer.
func writeDefault(b *bytes.Buffer, tag poTag, str string) error {
	var text strings.Builder

	nodes, err := messageformat.Parse(str)
	if err != nil {
		return fmt.Errorf("parse message: %w", err)
	}

	for _, node := range nodes {
		nodeTxt, ok := node.(messageformat.NodeText)
		if !ok {
			return errors.New("convert node to messageformat.NodeText")
		}

		text.WriteString(nodeTxt.Text)
	}

	if text.String() == "" {
		text.WriteString(str)
	}

	lines := getPoTagLines(text.String())

	if len(lines) == 1 {
		if _, err = fmt.Fprintf(b, "%s \"%s\"\n", tag, lines[0]); err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}

		return nil
	}

	if _, err = fmt.Fprintf(b, "%s \"\"\n", tag); err != nil {
		return fmt.Errorf("write %s: %w", tag, err)
	}

	if err = writeMultiline(b, tag, lines); err != nil {
		return fmt.Errorf("write multiline: %w", err)
	}

	return nil
}

// writePlural parses a plural message string into nodes, iterates over the nodes,
// and writes the variants of the plural message to a bytes.Buffer.
func writePlural(b *bytes.Buffer, tag poTag, str string) error {
	nodes, err := messageformat.Parse(str)
	if err != nil {
		return fmt.Errorf("parse message: %w", err)
	}

	for _, node := range nodes {
		nodeMatch, ok := node.(messageformat.NodeMatch)
		if !ok {
			return errors.New("convert node to messageformat.NodeMatch")
		}

		if err = writeVariants(b, tag, nodeMatch); err != nil {
			return fmt.Errorf("write variants: %w", err)
		}
	}

	return nil
}

// writeVariants writes the variants of a plural message to a bytes.Buffer.
func writeVariants(b *bytes.Buffer, tag poTag, nodeMatch messageformat.NodeMatch) error {
	for i, variant := range nodeMatch.Variants {
		if _, err := fmt.Fprintf(b, "msgstr[%d] ", i); err != nil {
			return fmt.Errorf("write plural msgstr: %w", err)
		}

		var txt strings.Builder

		for _, msg := range variant.Message {
			switch node := msg.(type) {
			case messageformat.NodeText:
				txt.WriteString(node.Text)
			case messageformat.NodeVariable:
				txt.WriteString("%d")
			default:
				return errors.New("unknown node type")
			}
		}

		lines := getPoTagLines(txt.String())

		if len(lines) == 1 {
			if _, err := fmt.Fprintf(b, "\"%s\"\n", lines[0]); err != nil {
				return fmt.Errorf("write %s: %w", tag, err)
			}

			continue
		}

		if _, err := fmt.Fprintf(b, "\"\"\n"); err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}

		if err := writeMultiline(b, tag, lines); err != nil {
			return fmt.Errorf("write multiline: %w", err)
		}
	}

	return nil
}

// writeMultiline writes a slice of strings as a multiline tag value to a bytes.Buffer.
func writeMultiline(b *bytes.Buffer, tag poTag, lines []string) error {
	for _, line := range lines {
		if !strings.HasSuffix(line, "\\n") {
			line += "\\n"
		}

		if _, err := fmt.Fprint(b, "\""+line+"\""+"\n"); err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}
	}

	return nil
}

// getPoTagLines processes a string by encoding, splitting it into lines,
// and handling special cases where the last line is an empty string due to trailing newline characters.
// It returns a slice of strings representing the individual lines.
func getPoTagLines(str string) []string {
	encodedStr := strconv.Quote(str)
	encodedStr = strings.ReplaceAll(encodedStr, "\\\\", "\\")
	encodedStr = encodedStr[1 : len(encodedStr)-1] // trim quotes
	lines := strings.Split(encodedStr, "\\n")

	// Remove the empty string element. The line is empty when splitting by "\\n" and "\n" is the last character in str.
	if lines[len(lines)-1] == "" {
		if len(lines) == 1 {
			return lines
		} else {
			lines = lines[:len(lines)-1]
			lines[len(lines)-1] += "\\n" // add the "\n" back to the last line
		}
	}

	return lines
}

// writeMessage writes a single message entry to a bytes.Buffer,
// including new lines, plural forms information, descriptions, fuzzy comment.
func writeMessage(b *bytes.Buffer, index int, message model.Message) error {
	if index > 0 {
		if _, err := fmt.Fprint(b, "\n"); err != nil {
			return fmt.Errorf("write new line: %w", err)
		}
	}

	if message.PluralID != "" {
		count := strings.Count(message.Message, "when")
		if _, err := fmt.Fprintf(b, "\"Plural-Forms: nplurals=%d; plural=(n != 1);\\n\"\n", count); err != nil {
			return fmt.Errorf("write plural forms: %w", err)
		}
	}

	descriptions := strings.Split(message.Description, "\n")

	for _, description := range descriptions {
		if description != "" {
			if _, err := fmt.Fprintf(b, "#. %s\n", description); err != nil {
				return fmt.Errorf("write description: %w", err)
			}
		}
	}

	for _, pos := range message.Positions {
		if _, err := fmt.Fprintf(b, "#: %s\n", pos); err != nil {
			return fmt.Errorf("write positions: %w", err)
		}
	}

	if message.Status == model.MessageStatusFuzzy {
		if _, err := fmt.Fprint(b, "#, fuzzy\n"); err != nil {
			return fmt.Errorf("write fuzzy: %w", err)
		}
	}

	return writeTags(b, message)
}

// writeTags writes specific tags (MsgId, MsgStr, PluralId, MsgStrPlural)
// along with their corresponding values to a bytes.Buffer.
func writeTags(b *bytes.Buffer, message model.Message) error {
	if err := writeToPoTag(b, MsgID, message.ID); err != nil {
		return fmt.Errorf("format msgid: %w", err)
	}

	// singular
	if message.PluralID == "" {
		if err := writeToPoTag(b, MsgStr, message.Message); err != nil {
			return fmt.Errorf("format msgstr: %w", err)
		}
	} else {
		// plural
		if err := writeToPoTag(b, PluralID, message.PluralID); err != nil {
			return fmt.Errorf("format msgid_plural: %w", err)
		}
		if err := writeToPoTag(b, MsgStrPlural, message.Message); err != nil {
			return fmt.Errorf("format msgstr[]: %w", err)
		}
	}

	return nil
}

// convertPluralsToMessageString converts a slice of strings to MessageFormat plural form.
func convertPluralsToMessageString(plurals []string) string {
	var sb strings.Builder

	sb.WriteString("match {$count :number}\n")

	for i, plural := range plurals {
		plural = escapeSpecialChars(plural)
		line := strings.ReplaceAll(strings.TrimSpace(plural), "%d", "{$count}")

		var count string

		if i == len(plurals)-1 {
			count = "*"
		} else {
			count = strconv.Itoa(i + 1)
		}

		sb.WriteString(fmt.Sprintf("when %s {%s}\n", count, line))
	}

	return sb.String()
}
