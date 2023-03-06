package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/pot"
)

type PoTag string

const (
	MsgId  PoTag = "msgid"
	MsgStr PoTag = "msgstr"
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

		if message.Description != "" {
			if _, err := fmt.Fprintf(&b, "#. %s\n", message.Description); err != nil {
				return nil, fmt.Errorf("write description: %w", err)
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
		if node.MsgIdPlural != "" {
			continue
		}

		fuzzy := strings.Contains(node.Flag, "fuzzy")

		message := model.Message{
			ID:          node.MsgId,
			Description: node.ExtractedComment,
			Fuzzy:       fuzzy,
		}

		if len(node.MsgStr) > 0 {
			message.Message = node.MsgStr[0]
		}

		messages = append(messages, message)
	}

	return model.Messages{
		Language: po.Header.Language,
		Messages: messages,
	}, nil
}

func writeToPoTag(b *bytes.Buffer, tag PoTag, str string) error {
	encodedStr, err := json.Marshal(str) // use JSON encoding to escape special characters
	if err != nil {
		return fmt.Errorf("marshal %s: %w", tag, err)
	}

	encodedStr = encodedStr[1 : len(encodedStr)-1] // trim quotes
	lines := strings.Split(string(encodedStr), "\\n")

	if len(lines) == 1 {
		_, err = fmt.Fprintf(b, "%s \"%s\"\n", tag, lines[0])
		if err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}

		return nil
	}

	multiline := make([]string, 0, len(lines))

	for i, line := range lines {
		if len(lines)-1 == i && line == "" {
			continue
		}

		multiline = append(multiline, "\""+line+"\\n\""+"\n")
	}

	_, err = fmt.Fprintf(b, "%s \"\"\n%s", tag, strings.Join(multiline, ""))
	if err != nil {
		return fmt.Errorf("write %s: %w", tag, err)
	}

	return nil
}
