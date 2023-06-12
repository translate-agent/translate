package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type modelType interface {
	model.Service
}

func Model[T modelType, V ~func(*T)](opts ...V) *T {
	var m T
	gofakeit.Struct(&m)

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

// ModelService returns a random model.Service.
func ModelService(opts ...ModelServiceOption) *model.Service {
	return Model(opts...)

	// service := &model.Service{
	// 	Name: gofakeit.FirstName(),
	// 	ID:   uuid.New(),
	// }

	// for _, opt := range opts {
	// 	opt(service)
	// }

	// return service
}

type ModelServiceOption func(*model.Service)

// WithName sets the name of the model.Service.
func WithName(name string) ModelServiceOption {
	return func(s *model.Service) {
		s.Name = name
	}
}

// WithID sets the ID of the model.Service.
func WithID(id uuid.UUID) ModelServiceOption {
	return func(s *model.Service) {
		s.ID = id
	}
}

// ModelMessage returns a random model.Message.
func ModelMessage() *model.Message {
	return &model.Message{
		ID:          gofakeit.SentenceSimple(),
		PluralID:    gofakeit.SentenceSimple(),
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

// WithoutMessages removes the messages (m.Messages[n].Message) from the model.Messages.
func WithoutMessages() ModelMessagesOption {
	return func(m *model.Messages) {
		for i := range m.Messages {
			m.Messages[i].Message = ""
			m.Messages[i].Fuzzy = false
		}
	}
}
