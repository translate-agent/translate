package convert

import (
	"encoding/xml"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type xmb struct {
	XMLName  xml.Name     `xml:"messagebundle"`
	Language language.Tag `xml:"srcLang,attr"`
	Messages []msg        `xml:"message"`
}

type msg struct {
	ID          string `xml:"id,attr"`
	Description string `xml:"desc,attr"`
	Content     string `xml:"content"`
	Fuzzy       bool   `xml:"fuzzy,attr"`
}

func FromXMB(b []byte) (model.Messages, error) {
	var xmbFormat xmb

	err := xml.Unmarshal(b, &xmbFormat)
	if err != nil {
		return model.Messages{}, fmt.Errorf("unmarshaling xmb to model.Messages: %w", err)
	}

	messages := model.Messages{Messages: make([]model.Message, 0, len(xmbFormat.Messages))}
	for _, message := range xmbFormat.Messages {
		messages.Messages = append(messages.Messages, model.Message{
			ID:          message.ID,
			Message:     message.Content,
			Description: message.Description,
			Fuzzy:       message.Fuzzy,
		})
	}

	return model.Messages{
		Language: xmbFormat.Language,
		Messages: messages.Messages,
	}, nil
}

func ToXMB(m model.Messages) ([]byte, error) {
	xmbFormat := xmb{
		Messages: make([]msg, len(m.Messages)),
		Language: m.Language,
	}

	for i, message := range m.Messages {
		xmbFormat.Messages[i] = msg{
			ID:          message.ID,
			Content:     message.Message,
			Description: message.Description,
			Fuzzy:       message.Fuzzy,
		}
	}

	data, err := xml.Marshal(&xmbFormat)
	if err != nil {
		return nil, fmt.Errorf("marshaling xmb to model.Messages: %w", err)
	}

	dataWithHeader := append([]byte(xml.Header), data...)

	return dataWithHeader, nil
}
