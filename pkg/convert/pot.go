package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.expect.digital/translate/pkg/pot"

	"go.expect.digital/translate/pkg/model"
)

type poTag string

const (
	MsgId        poTag = "msgid"
	PluralId     poTag = "msgid_plural"
	MsgStrPlural poTag = "msgstr[%d]"
	MsgStr       poTag = "msgstr"
)

func ToPot(m model.Messages) ([]byte, error) {
	var b bytes.Buffer

	if _, err := fmt.Fprintf(&b, "\"Language: %s\n", m.Language); err != nil {
		return nil, fmt.Errorf("write language: %w", err)
	}

	for i, message := range m.Messages {
		if err := writeMessage(&b, i, message); err != nil {
			return nil, fmt.Errorf("write message: %w", err)
		}
	}

	return b.Bytes(), nil
}

func FromPot(b []byte) (model.Messages, error) {
	pluralCountLimit := 2

	tokens, err := pot.Lex(bytes.NewReader(b))
	if err != nil {
		return model.Messages{}, fmt.Errorf("dividing po file to tokens: %w", err)
	}

	po, err := pot.TokensToPo(tokens)
	if err != nil {
		return model.Messages{}, fmt.Errorf("convert tokens to pot.Po: %w", err)
	}

	messages := make([]model.Message, 0, len(po.Messages))

	for _, node := range po.Messages {
		message := model.Message{ID: node.MsgId}

		switch {
		case po.Header.PluralForms.NPlurals > pluralCountLimit:
			return model.Messages{}, fmt.Errorf("plural forms with more than 2 forms are not implemented yet")
		case po.Header.PluralForms.NPlurals > 1:
			message.PluralID = node.MsgIdPlural
			message.Message = convertToFormatMessage(node.MsgStr)
		default:
			message.Message = node.MsgStr[0]
		}

		message.Description = strings.Join(node.ExtractedComment, "\n ")
		message.Fuzzy = strings.Contains(node.Flag, "fuzzy")

		messages = append(messages, message)
	}

	return model.Messages{
		Language: po.Header.Language,
		Messages: messages,
	}, nil
}

func writeToPoTag(b *bytes.Buffer, tag poTag, str string) error {
	var write func(*bytes.Buffer, poTag, string) error

	switch tag { //nolint:exhaustive
	case MsgStrPlural:
		write = writePluralForms
	default:
		write = writeDefault
	}

	if err := write(b, tag, str); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func writeDefault(b *bytes.Buffer, tag poTag, str string) error {
	lines, err := getPoTagLines(str, tag)
	if err != nil {
		return fmt.Errorf("get po tag lines: %w", err)
	}

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

func writePluralForms(b *bytes.Buffer, tag poTag, str string) error {
	pluralForms, err := extractPluralForms(str)
	if err != nil {
		return fmt.Errorf("extract plural forms: %w", err)
	}

	for i := range pluralForms {
		pluralLines, err := getPoTagLines(pluralForms[i], tag)
		if err != nil {
			return fmt.Errorf("get po tag lines: %w", err)
		}

		if len(pluralLines) == 1 {
			if _, err = fmt.Fprintf(b, "msgstr[%d] \"%s\"\n", i, pluralLines[0]); err != nil {
				return fmt.Errorf("write %s: %w", tag, err)
			}

			continue
		}

		if _, err = fmt.Fprintf(b, "msgstr[%d] \"\"\n", i); err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}

		if err = writeMultiline(b, tag, pluralLines); err != nil {
			return fmt.Errorf("write plural multiline: %w", err)
		}
	}

	return nil
}

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

func getPoTagLines(str string, tag poTag) ([]string, error) {
	encodedStr, err := json.Marshal(str) // use JSON encoding to escape special characters
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", tag, err)
	}

	encodedStr = encodedStr[1 : len(encodedStr)-1] // trim quotes
	lines := strings.Split(string(encodedStr), "\\n")

	// Remove the empty string element. The line is empty when splitting by "\\n" and "\n" is the last character in str.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
		lines[len(lines)-1] += "\\n" // add the "\n" back to the last line
	}

	return lines, nil
}

func extractPluralForms(plural string) ([]string, error) {
	if plural == "" {
		return nil, fmt.Errorf("plural is empty")
	}

	plural = strings.ReplaceAll(plural, "{$count}", "%d")
	// Use regular expressions to extract the plural forms from the plural finds string in {}
	re := regexp.MustCompile(`when [\d*]+ \{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(plural, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no plural forms found in plural")
	}

	// Create a slice of strings to hold the extracted plural forms
	pluralForms := make([]string, 0, len(matches))

	// Loop through the matches and extract the plural form from each one
	for _, match := range matches {
		pluralForms = append(pluralForms, match[1])
	}

	return pluralForms, nil
}

func convertToFormatMessage(plurals []string) string {
	var sb strings.Builder

	sb.WriteString("match {$count :number}\n")

	for i, plural := range plurals {
		line := strings.ReplaceAll(strings.TrimSpace(plural), "%d", "{$count}")

		var count string

		switch i {
		default:
			count = fmt.Sprintf("%d", i+1)
		case len(plurals) - 1:
			count = "*"
		}

		sb.WriteString(fmt.Sprintf("when %s {%s}\n", count, line))
	}

	return sb.String()
}

func writeMessage(b *bytes.Buffer, index int, message model.Message) error {
	if index > 0 {
		if _, err := fmt.Fprint(b, "\n"); err != nil {
			return fmt.Errorf("write new line: %w", err)
		}
	}

	if message.PluralID != "" {
		count := strings.Count(message.Message, "when")
		if _, err := fmt.Fprintf(b, "# Plural-Forms: nplurals=%d; plural=(n != 1);\n", count); err != nil {
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

	if message.Fuzzy {
		if _, err := fmt.Fprint(b, "#, fuzzy\n"); err != nil {
			return fmt.Errorf("write fuzzy: %w", err)
		}
	}

	if err := writeTags(b, message); err != nil {
		return err
	}

	return nil
}

func writeTags(b *bytes.Buffer, message model.Message) error {
	if err := writeToPoTag(b, MsgId, message.ID); err != nil {
		return fmt.Errorf("format msgid: %w", err)
	}

	switch {
	case message.PluralID != "":
		if err := writeToPoTag(b, PluralId, message.PluralID); err != nil {
			return fmt.Errorf("format msgid_plural: %w", err)
		}

		if err := writeToPoTag(b, MsgStrPlural, message.Message); err != nil {
			return fmt.Errorf("format msgstr[]: %w", err)
		}
	default:
		if err := writeToPoTag(b, MsgStr, message.Message); err != nil {
			return fmt.Errorf("format msgstr: %w", err)
		}
	}

	return nil
}
