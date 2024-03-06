package po

import (
	"regexp"
	"strings"
)

type parser struct {
	lines []string
	pos   int
}

const eof = "eof"

func (p *parser) peek() string {
	next := p.next()
	p.pos--

	return next
}

func (p *parser) next() string {
	if p.pos+1 >= len(p.lines) {
		return eof
	}

	p.pos++

	return strings.TrimSpace(p.lines[p.pos])
}

func (p *parser) parseHead() Headers {
	// hasHeaders checks if the file has headers, and if so sets pos to the first header line.
	hasHeaders := func() bool {
		for line := p.next(); line != eof; line = p.next() {
			if line == `msgid ""` && p.peek() == `msgstr ""` {
				p.pos++
				return true
			}
		}

		p.pos = -1 // reset the position

		return false
	}

	if !hasHeaders() {
		return nil
	}

	var buff string                                                    // buffer for headers
	for line := p.next(); line != "" && line != eof; line = p.next() { // until next newline
		buff += line + "\n"
	}

	return p.emitHeaders(buff)
}

func (p *parser) emitHeaders(buff string) Headers {
	re := regexp.MustCompile(`"(?s)([A-Za-z-]+):\s(.*?)\\n"`)
	// Explanation:
	// "(?s)        - match the quote and enable the dot to match newlines (for multiline headers)
	// ([A-Za-z-]+) - match the header name, which is a sequence of letters and hyphens
	// :\s          - match the colon and the space after the header name
	// (.*?)        - match the header value, which could be any character
	// \\n"         - match the newline and the quote at the end of the header value
	matches := re.FindAllStringSubmatch(buff, -1)

	if matches == nil {
		return nil
	}

	headers := make(Headers, 0, len(matches))
	for _, match := range matches {
		headers = append(headers, Header{
			Name:  match[1],
			Value: strings.NewReplacer(`"`, "").Replace(match[2]),
		})
	}

	return headers
}

func (p *parser) parseMessages() []Message {
	var messages []Message
	for p.peek() != eof {
		messages = append(messages, p.parseMessage())
	}

	return messages
}

type state int

// state used to track the last state to handle multiline strings.
const (
	msgID state = iota
	msgIDPlural
	msgStr
)

func (p *parser) parseMessage() Message {
	var (
		msg       Message
		lastState state // track the last state to handle multiline strings
	)

	for line := p.next(); line != "" && line != eof; line = p.next() {
		switch {
		case strings.HasPrefix(line, "# "):
			msg.TranslatorComments = append(msg.TranslatorComments, line[2:])
		case strings.HasPrefix(line, "#. "):
			msg.ExtractedComments = append(msg.ExtractedComments, line[3:])
		case strings.HasPrefix(line, "#: "):
			msg.References = append(msg.References, line[3:])
		case strings.HasPrefix(line, `#, `):
			msg.Flags = append(msg.Flags, line[3:])
		case strings.HasPrefix(line, `msgid "`):
			lastState = msgID
			msg.MsgID = line[7 : len(line)-1]
		case strings.HasPrefix(line, `msgstr "`):
			lastState = msgStr

			msg.MsgStr = append(msg.MsgStr, line[8:len(line)-1])
		case strings.HasPrefix(line, `msgstr[`):
			lastState = msgStr
			idx := strings.Index(line, `] "`)
			msg.MsgStr = append(msg.MsgStr, line[idx+3:len(line)-1])
		case strings.HasPrefix(line, `msgid_plural "`):
			lastState = msgIDPlural
			msg.MsgIDPlural = line[14 : len(line)-1]
		case strings.HasPrefix(line, `"`):
			lineVal := line[1 : len(line)-1]
			if strings.HasSuffix(line, `\n"`) {
				lineVal = line[1:len(line)-3] + "\n"
			}

			switch lastState {
			case msgID:
				msg.MsgID += lineVal
			case msgIDPlural:
				msg.MsgIDPlural += lineVal
			case msgStr:
				msg.MsgStr[len(msg.MsgStr)-1] += lineVal
			}
		}
	}

	// remove the empty string if it's the only element
	if len(msg.MsgStr) == 1 && msg.MsgStr[0] == "" {
		msg.MsgStr = []string{}
	}

	return msg
}

// Parse parses the input and returns a PO struct representing the gettext's Portable Object file.
func Parse(input []byte) (PO, error) {
	p := parser{
		lines: strings.Split(string(input), "\n"),
		pos:   -1,
	}

	return PO{Headers: p.parseHead(), Messages: p.parseMessages()}, nil
}
