package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type modelType interface {
	model.Service | model.Message | model.Messages
}

// slice returns a slice of random modelType.
func slice[T modelType, V ~func(*T)](count uint, f func(opts ...V) *T, opts ...V) []*T {
	var ms []*T
	for i := uint(0); i < count; i++ {
		ms = append(ms, f(opts...))
	}

	return ms
}

// ------------------Service------------------

// ModelService returns a random model.Service.
func ModelService(opts ...ModelServiceOption) *model.Service {
	service := model.Service{
		Name: gofakeit.Name(),
		ID:   uuid.New(),
	}

	for _, opt := range opts {
		opt(&service)
	}

	return &service
}

// ModelServiceSlice returns a slice of random model.Service.
func ModelServiceSlice(count uint, opts ...ModelServiceOption) []*model.Service {
	return slice(count, ModelService, opts...)
}

// ------------------Service Opts------------------

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

// ------------------Message------------------

// ModelMessage returns a random model.Message.
func ModelMessage(opts ...ModelMessageOption) *model.Message {
	opts = append(opts, WithEmptyPluralID()) // PluralID is ignored as it is not used in the repo yet

	msg := model.Message{
		ID:          gofakeit.SentenceSimple(),
		PluralID:    gofakeit.SentenceSimple(),
		Message:     gofakeit.SentenceSimple(),
		Description: gofakeit.SentenceSimple(),
		Fuzzy:       gofakeit.Bool(),
	}

	for _, opt := range opts {
		opt(&msg)
	}

	return &msg
}

// ModelMessageSlice returns a slice of random model.Message.
func ModelMessageSlice(count uint, opts ...ModelMessageOption) []*model.Message {
	return slice(count, ModelMessage, opts...)
}

// ------------------Message Opts------------------

type ModelMessageOption func(*model.Message)

func WithEmptyMessage() ModelMessageOption {
	return func(m *model.Message) {
		m.Message = ""
		m.Fuzzy = false
	}
}

func WithEmptyPluralID() ModelMessageOption {
	return func(m *model.Message) {
		m.PluralID = ""
	}
}

// ------------------Messages------------------

// ModelMessages returns a random model.Messages.
func ModelMessages(count uint, msgOpts []ModelMessageOption, opts ...ModelMessagesOption) *model.Messages {
	messages := &model.Messages{Language: Lang()}
	if count > 0 {
		messages.Messages = make([]model.Message, count)
	}

	msgs := ModelMessageSlice(count, msgOpts...)
	for i, msg := range msgs {
		messages.Messages[i] = *msg
	}

	for _, opt := range opts {
		opt(messages)
	}

	return messages
}

func ModelMessagesSlice(
	count uint,
	diffLang bool,
	msgOpts []ModelMessageOption,
	opts ...ModelMessagesOption,
) []*model.Messages {
	modelMsg := func(opts ...ModelMessagesOption) *model.Messages {
		return ModelMessages(3, msgOpts, opts...) //nolint:gomnd
	}

	msgsSlice := slice(count, modelMsg, opts...)

	if !diffLang {
		return msgsSlice
	}

	// Assign different languages to each messages
	for i, lang := range Langs(count) {
		msgsSlice[i].Language = lang
	}

	return msgsSlice
}

// ------------------Messages Opts------------------

type ModelMessagesOption func(*model.Messages)

// WithLanguage sets the language of the model.Messages.
func WithLanguage(lang language.Tag) ModelMessagesOption {
	return func(m *model.Messages) {
		m.Language = lang
	}
}
