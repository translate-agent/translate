package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Message() *model.Message {
	return &model.Message{
		ID:          gofakeit.SentenceSimple(),
		Message:     gofakeit.SentenceSimple(),
		Description: gofakeit.SentenceSimple(),
		Fuzzy:       gofakeit.Bool(),
	}
}

func Messages(count uint, opts ...MessagesOption) *model.Messages {
	msgs := make([]model.Message, 0, count)
	for i := uint(0); i < count; i++ {
		msgs = append(msgs, *Message())
	}

	messages := &model.Messages{Language: Lang(), Messages: msgs}

	for _, opt := range opts {
		opt(messages)
	}

	return messages
}

type MessagesOption func(*model.Messages)

func WithLanguage(lang language.Tag) MessagesOption {
	return func(m *model.Messages) {
		m.Language = lang
	}
}

func WithoutTranslations() MessagesOption {
	return func(m *model.Messages) {
		for i := range m.Messages {
			m.Messages[i].Message = ""
			m.Messages[i].Fuzzy = false
		}
	}
}
