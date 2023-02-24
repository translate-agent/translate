package convert

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/pot"
)

func ToPot(m model.Messages) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("\"Language: %s\n", m.Language))

	for _, message := range m.Messages {
		if message.Description != "" {
			_, err := fmt.Fprintf(&b, "#. %s\n", message.Description)
			if err != nil {
				return nil, fmt.Errorf("write description: %w", err)
			}
		}

		if message.Fuzzy {
			_, err := fmt.Fprint(&b, "#, fuzzy\n")
			if err != nil {
				return nil, fmt.Errorf("write fuzzy: %w", err)
			}
		}

		if strings.HasSuffix(message.ID, "\\n") {
			_, err := fmt.Fprintf(&b, "msgid \"\" \n\"%s\"\n", message.ID)
			if err != nil {
				return nil, fmt.Errorf("write msgid: %w", err)
			}
		} else {
			messageIdWithQuotes := strings.ReplaceAll(message.ID, "\"", "\\\"")
			_, err := fmt.Fprintf(&b, "msgid \"%s\"\n", messageIdWithQuotes)
			if err != nil {
				return nil, fmt.Errorf("write msgid: %w", err)
			}
		}

		if strings.HasSuffix(message.Message, "\\n") {
			_, err := fmt.Fprintf(&b, "msgstr \"\" \n\"%s\"\n", message.Message)
			if err != nil {
				return nil, fmt.Errorf("write msgid: %w", err)
			}
		} else {
			messageWithQuotes := strings.ReplaceAll(message.Message, "\"", "\\\"")
			_, err := fmt.Fprintf(&b, "msgstr \"%s\"\n", messageWithQuotes)
			if err != nil {
				return nil, fmt.Errorf("write msgid: %w", err)
			}
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
