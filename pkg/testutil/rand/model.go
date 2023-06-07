package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"go.expect.digital/translate/pkg/model"
)

func Message(opts ...MessageOption) model.Message {
	msgs := &model.Message{
		ID:          gofakeit.SentenceSimple(),
		Message:     gofakeit.SentenceSimple(),
		Description: gofakeit.SentenceSimple(),
		Fuzzy:       gofakeit.Bool(),
	}

	for _, opt := range opts {
		opt(msgs)
	}

	return *msgs
}

type MessageOption func(*model.Message)

func WithoutTranslation() MessageOption {
	return func(m *model.Message) {
		m.Message = ""
		m.Fuzzy = false
	}
}

func Messages(count uint, opts ...MessageOption) *model.Messages {
	messages := make([]model.Message, 0, count)
	for i := uint(0); i < count; i++ {
		messages = append(messages, Message(opts...))
	}

	return &model.Messages{Language: Lang(), Messages: messages}
}
