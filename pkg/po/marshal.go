package po

import (
	"bytes"
	"fmt"
	"strings"
)

// Marshal serializes the Po object into a byte slice.
func (p *PO) Marshal() []byte {
	var b bytes.Buffer

	p.Headers.marshal(&b)

	for _, msg := range p.Messages {
		msg.marshal(&b)
	}

	return b.Bytes()
}

func (m *Message) marshal(b *bytes.Buffer) {
	// quoteLines function splits a string into multiple lines and wraps each line in double quotes.
	quoteLines := func(s string) string {
		split := strings.Split(s, "\n")
		for i, line := range split {
			split[i] = fmt.Sprintf("\"%s\"", line)
		}

		return strings.Join(split, "\n")
	}

	if b.Len() > 0 {
		b.WriteRune('\n') // empty line before each message, except the first one, if headers are not present.
	}

	for _, translatorComment := range m.TranslatorComments {
		b.WriteString(fmt.Sprintf("# %s\n", translatorComment))
	}

	for _, extractedComment := range m.ExtractedComments {
		b.WriteString(fmt.Sprintf("#. %s\n", extractedComment))
	}

	for _, reference := range m.References {
		b.WriteString(fmt.Sprintf("#: %s\n", reference))
	}

	for _, flag := range m.Flags {
		b.WriteString(fmt.Sprintf("#, %s\n", flag))
	}

	if m.MsgID != "" {
		b.WriteString(fmt.Sprintf("msgid %s\n", quoteLines(m.MsgID)))
	}

	if m.MsgIDPlural != "" {
		b.WriteString(fmt.Sprintf("msgid_plural %s\n", quoteLines(m.MsgIDPlural)))
	}

	switch len(m.MsgStr) {
	case 0: // empty
		b.WriteString("msgstr \"\"\n")
	case 1: // singular
		b.WriteString(fmt.Sprintf("msgstr %s\n", quoteLines(m.MsgStr[0])))
	default: // plural
		for i, ms := range m.MsgStr {
			b.WriteString(fmt.Sprintf("msgstr[%d] %s\n", i, quoteLines(ms)))
		}
	}
}

func (h Headers) marshal(b *bytes.Buffer) {
	if len(h) == 0 {
		return
	}

	b.WriteString("msgid \"\"\nmsgstr \"\"\n")

	for _, header := range h {
		b.WriteString(fmt.Sprintf("\"%s: %s\\n\"\n", header.Name, header.Value))
	}
}
