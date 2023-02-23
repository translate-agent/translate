package convert

import (
	"bytes"
	"encoding/json"
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
			b.WriteString(fmt.Sprintf("#. %s\n", message.Description))
		}

		if message.Fuzzy {
			b.WriteString("#, fuzzy\n")
		}

		b.WriteString(fmt.Sprintf("msgid \"%s\"\n", message.ID))

		b.WriteString(fmt.Sprintf("msgstr \"%s\"\n", message.Message))
	}

	return b.Bytes(), nil
}

func FromPot(b []byte) (model.Messages, error) {
	var po *pot.Po
	if err := json.Unmarshal(b, &po); err != nil {
		return model.Messages{}, fmt.Errorf("error unmarshaling byte slice to pot.Po: %w", err)
	}

	messages := make([]model.Message, 0, len(po.Messages))

	for _, node := range po.Messages {
		if node.MsgIdPlural != "" {
			continue
		}

		var fuzzy bool
		if strings.Contains(node.Flag, "fuzzy") {
			fuzzy = true
		}

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
