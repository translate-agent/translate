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

		msgId, err := formatMsgId(message.ID)
		if err != nil {
			return nil, fmt.Errorf("format msgid: %w", err)
		}

		if _, err = fmt.Fprint(&b, msgId); err != nil {
			return nil, fmt.Errorf("write msgid: %w", err)
		}

		msgStr, err := formatMsgStr(message.Message)
		if err != nil {
			return nil, fmt.Errorf("format msgstr: %w", err)
		}

		if _, err = fmt.Fprint(&b, msgStr); err != nil {
			return nil, fmt.Errorf("write msgstr: %w", err)
		}
	}

	return b.Bytes(), nil
}

func FromPot(b []byte) (model.Messages, error) {
	tokens := pot.Lex(bufio.NewReader(bytes.NewReader(b)))

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

func formatMsgId(msgID string) (string, error) {
	msgIDBytes, err := json.Marshal(msgID)
	if err != nil {
		return "", fmt.Errorf("marshal msgid: %w", err)
	}

	msgIdTextLines := strings.Split(string(msgIDBytes), "\\n")

	if len(msgIdTextLines) == 1 {
		return fmt.Sprintf("msgid %s\n", msgIdTextLines[0]), nil
	}

	lines := make([]string, 0, len(msgIdTextLines))

	for i, line := range msgIdTextLines {
		if len(msgIdTextLines)-1 == i && line == "\"" {
			continue
		}

		line = strings.ReplaceAll(line, "\"", "")
		lines = append(lines, "\""+line+"\\n\""+"\n")
	}

	return fmt.Sprintf("msgid \"\"\n%s", strings.Join(lines, "")), nil
}

func formatMsgStr(msgStr string) (string, error) {
	msgStrBytes, err := json.Marshal(msgStr)
	if err != nil {
		return "", fmt.Errorf("marshal msgstr: %w", err)
	}

	msgStrTextLines := strings.Split(string(msgStrBytes), "\\n")

	if len(msgStrTextLines) == 1 {
		return fmt.Sprintf("msgstr %s\n\n", msgStrTextLines[0]), nil
	}

	lines := make([]string, 0, len(msgStrTextLines))

	for i, line := range msgStrTextLines {
		if len(msgStrTextLines)-1 == i && line == "\"" {
			continue
		}

		line = strings.ReplaceAll(line, "\"", "")
		lines = append(lines, "\""+line+"\\n\""+"\n")
	}

	return fmt.Sprintf("msgstr \"\"\n%s\n", strings.Join(lines, "")), nil
}
