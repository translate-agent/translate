package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/pot"
)

type poTag string

const (
	MsgId  poTag = "msgid"
	MsgStr poTag = "msgstr"
)

func ToPot(m model.Messages) ([]byte, error) {
	var b bytes.Buffer

	if _, err := fmt.Fprintf(&b, "\"Language: %s\n", m.Language); err != nil {
		return nil, fmt.Errorf("write language: %w", err)
	}

	for i, message := range m.Messages {
		if i > 0 {
			if _, err := fmt.Fprint(&b, "\n"); err != nil {
				return nil, fmt.Errorf("write new line: %w", err)
			}
		}

		descriptions := strings.Split(message.Description, "\n")

		for _, description := range descriptions {
			if description != "" {
				if _, err := fmt.Fprintf(&b, "#. %s\n", description); err != nil {
					return nil, fmt.Errorf("write description: %w", err)
				}
			}
		}

		if message.Fuzzy {
			if _, err := fmt.Fprint(&b, "#, fuzzy\n"); err != nil {
				return nil, fmt.Errorf("write fuzzy: %w", err)
			}
		}

		err := writeToPoTag(&b, MsgId, message.ID)
		if err != nil {
			return nil, fmt.Errorf("format msgid: %w", err)
		}

		err = writeToPoTag(&b, MsgStr, message.Message)
		if err != nil {
			return nil, fmt.Errorf("format msgstr: %w", err)
		}
	}

	return b.Bytes(), nil
}

func FromPot(b []byte) (model.Messages, error) {
	tokens, err := pot.Lex(bytes.NewReader(b))
	if err != nil {
		return model.Messages{}, fmt.Errorf("dividing po file to tokens: %w", err)
	}

	po, err := pot.TokensToPo(tokens)
	if err != nil {
		return model.Messages{}, fmt.Errorf("convert tokens to pot.Po: %w", err)
	}

	messages := make([]model.Message, 0, len(po.Messages))
	// model.Messages does not support plurals atm. But we plan to impl it under:
	// https://github.com/orgs/translate-agent/projects/1?pane=issue&itemId=21251425
	for _, node := range po.Messages {
		if strings.HasSuffix(node.MsgId, "\\n") {
			node.MsgId = strings.ReplaceAll(node.MsgId, "\\n", "\n")
		}

		if node.MsgIdPlural != "" {
			continue
		}

		fuzzy := strings.Contains(node.Flag, "fuzzy")

		message := model.Message{
			ID:          node.MsgId,
			Description: strings.Join(node.ExtractedComment, "\n"),
			Fuzzy:       fuzzy,
		}

		if len(node.MsgStr) > 0 {
			if strings.HasSuffix(node.MsgStr[0], "\\n") {
				node.MsgStr[0] = strings.ReplaceAll(node.MsgStr[0], "\\n", "\n")
			}

			message.Message = node.MsgStr[0]
		}

		messages = append(messages, message)
	}

	return model.Messages{
		Language: po.Header.Language,
		Messages: messages,
	}, nil
}

func writeToPoTag(b *bytes.Buffer, tag poTag, str string) error {
	encodedStr, err := json.Marshal(str) // use JSON encoding to escape special characters
	if err != nil {
		return fmt.Errorf("marshal %s: %w", tag, err)
	}

	encodedStr = encodedStr[1 : len(encodedStr)-1] // trim quotes
	lines := strings.Split(string(encodedStr), "\\n")

	// Remove the empty string element. The line is empty when splitting by "\\n" and "\n" is the last character in str.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
		lines[len(lines)-1] = lines[len(lines)-1] + "\\n" // add the "\n" back to the last line
	}

	if len(lines) == 1 {
		_, err = fmt.Fprintf(b, "%s \"%s\"\n", tag, lines[0])
		if err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}

		return nil
	}

	_, err = fmt.Fprintf(b, "%s \"\"\n", tag)
	if err != nil {
		return fmt.Errorf("write %s: %w", tag, err)
	}

	for _, line := range lines {
		if strings.HasSuffix(line, "\\n") {
			_, err = fmt.Fprint(b, "\""+line+"\""+"\n")
			if err != nil {
				return fmt.Errorf("write %s: %w", tag, err)
			}
		} else {
			_, err = fmt.Fprint(b, "\""+line+"\\n\""+"\n")
			if err != nil {
				return fmt.Errorf("write %s: %w", tag, err)
			}
		}
	}

	return nil
}
