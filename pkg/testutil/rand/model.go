package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// ModelMessage returns a random model.Message.
func ModelMessage() *model.Message {
	return &model.Message{
		ID:          gofakeit.SentenceSimple(),
		Message:     gofakeit.SentenceSimple(),
		Description: gofakeit.SentenceSimple(),
		Fuzzy:       gofakeit.Bool(),
	}
}

// ModelMessages returns a random model.Messages.
func ModelMessages(count uint, opts ...ModelMessagesOption) *model.Messages {
	msgs := make([]model.Message, 0, count)
	for i := uint(0); i < count; i++ {
		msgs = append(msgs, *ModelMessage())
	}

	messages := &model.Messages{Language: Lang(), Messages: msgs}

	for _, opt := range opts {
		opt(messages)
	}

	return messages
}

type ModelMessagesOption func(*model.Messages)

// WithLanguage sets the language of the model.Messages.
func WithLanguage(lang language.Tag) ModelMessagesOption {
	return func(m *model.Messages) {
		m.Language = lang
	}
}

// WithoutTranslations removes the translations (m.Messages[n].Message) from the model.Messages.
func WithoutTranslations() ModelMessagesOption {
	return func(m *model.Messages) {
		for i := range m.Messages {
			m.Messages[i].Message = ""
			m.Messages[i].Fuzzy = false
		}
	}
}
