package convert

import (
	"bufio"
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

	for _, message := range m.Messages {
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

		msgId, err := formatPoTags(message.ID, MsgId)
		if err != nil {
			return nil, fmt.Errorf("format msgid: %w", err)
		}

		if _, err = fmt.Fprint(&b, msgId); err != nil {
			return nil, fmt.Errorf("write msgid: %w", err)
		}

		msgStr, err := formatPoTags(message.Message, MsgStr)
		if err != nil {
			return nil, fmt.Errorf("format msgstr: %w", err)
		}

		if _, err = fmt.Fprint(&b, msgStr); err != nil {
			return nil, fmt.Errorf("write msgstr: %w", err)
		}

		if _, err = fmt.Fprint(&b, "\n"); err != nil {
			return nil, fmt.Errorf("write new line: %w", err)
		}
	}

	return b.Bytes(), nil
}

func FromPot(b []byte) (model.Messages, error) {
	tokens, err := pot.Lex(bufio.NewReader(bytes.NewReader(b)))
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

func formatPoTags(str string, poTag PoTag) (string, error) {
	b, err := json.Marshal(str)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", poTag, err)
	}

	b = b[1 : len(b)-1]
	lines := strings.Split(string(b), "\\n")

	if len(lines) == 1 {
		return fmt.Sprintf("%s \"%s\"\n", poTag, lines[0]), nil
	}

	multiline := make([]string, 0, len(lines))

	for i, line := range lines {
		if len(lines)-1 == i && line == "" {
			continue
		}

		line = strings.ReplaceAll(line, "\"", "")
		multiline = append(multiline, "\""+line+"\\n\""+"\n")
	}

	return fmt.Sprintf("%s \"\"\n%s", poTag, strings.Join(multiline, "")), nil
}
