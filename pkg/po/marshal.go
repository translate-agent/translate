package po

import (
	"bytes"
	"fmt"
	"strconv"
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
	// writeQuoted function splits a string into multiple lines and wraps each line in double quotes.
	writeQuoted := func(s string) {
		b.WriteRune('"')

		for _, r := range s {
			switch r {
			case '\n':
				b.WriteString("\"\n\"")
				continue
			case '"':
				b.WriteRune('\\')
			}

			b.WriteRune(r)
		}

		b.WriteString("\"\n")
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
		b.WriteString("msgid ")
		writeQuoted(m.MsgID)
	}

	if m.MsgIDPlural != "" {
		b.WriteString("msgid_plural ")
		writeQuoted(m.MsgIDPlural)
	}

	switch len(m.MsgStr) {
	case 0: // empty
		b.WriteString("msgstr \"\"\n")
	case 1: // singular
		b.WriteString("msgstr ")
		writeQuoted(m.MsgStr[0])
	default: // plural
		for i, ms := range m.MsgStr {
			b.WriteString("msgstr[" + strconv.Itoa(i) + "] ")
			writeQuoted(ms)
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
