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

// mdl returns a random modelType. It uses the given randF to generate the modelType.
func mdl[T modelType, O ~func(*T)](randF func() *T, opts ...O) *T {
	m := randF()

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// slice returns a slice of random modelType.
func slice[T modelType, O ~func(*T)](n uint, f func(opts ...O) *T, opts ...O) []*T {
	var s []*T
	for i := uint(0); i < n; i++ {
		s = append(s, f(opts...))
	}

	return s
}

// ------------------Service------------------

// modelService generates a random model.Service.
func modelService() *model.Service {
	return &model.Service{
		Name: gofakeit.Name(),
		ID:   uuid.New(),
	}
}

// ModelService generates a random model.Service using provided options.
func ModelService(opts ...ModelServiceOption) *model.Service {
	return mdl(modelService, opts...)
}

// ModelServiceSlice generates a slice of random model.Service using the provided options for each service.
func ModelServiceSlice(n uint, opts ...ModelServiceOption) []*model.Service {
	return slice(n, ModelService, opts...)
}

// ------------------Service Opts------------------

type ModelServiceOption func(*model.Service)

// WithID sets the ID of the model.Service.
func WithID(id uuid.UUID) ModelServiceOption {
	return func(s *model.Service) {
		s.ID = id
	}
}

// ------------------Message------------------

// modelMessage generates a random model.Message.
func modelMessage() *model.Message {
	return &model.Message{
		ID:          gofakeit.SentenceSimple(),
		PluralID:    gofakeit.SentenceSimple(),
		Message:     gofakeit.SentenceSimple(),
		Description: gofakeit.SentenceSimple(),
		Fuzzy:       gofakeit.Bool(),
	}
}

// ModelMessage generates a random model.Message using provided options.
func ModelMessage(opts ...ModelMessageOption) *model.Message {
	opts = append(opts, WithEmptyPluralID()) // PluralID is ignored as it is not used in the repo yet

	return mdl(modelMessage, opts...)
}

// ModelMessageSlice generates a slice of random model.Message using the provided options for each message.
func ModelMessageSlice(n uint, opts ...ModelMessageOption) []*model.Message {
	return slice(n, ModelMessage, opts...)
}

// ------------------Message Opts------------------

type ModelMessageOption func(*model.Message)

// WithEmptyPluralID sets the PluralID of the model.Message to "".
func WithEmptyPluralID() ModelMessageOption {
	return func(m *model.Message) {
		m.PluralID = ""
	}
}

// ------------------Messages------------------

// modelMessages generates a random model.Messages with the given
// count of Messages.messages and using the provided options for each message.
func modelMessages(msgCount uint, msgOpts ...ModelMessageOption) *model.Messages {
	messages := &model.Messages{Language: Language(), Messages: make([]model.Message, msgCount)}

	if msgCount == 0 {
		return messages
	}

	msgs := ModelMessageSlice(msgCount, msgOpts...)
	for i, msg := range msgs {
		messages.Messages[i] = *msg
	}

	return messages
}

// ModelMessages generates a random model.Messages with specific messages.message count, message and messages options.
func ModelMessages(msgCount uint, msgOpts []ModelMessageOption, msgsOpts ...ModelMessagesOption) *model.Messages {
	// msgsF wraps modelMessages() for mdl function.
	msgsF := func() *model.Messages {
		return modelMessages(msgCount, msgOpts...)
	}

	return mdl(msgsF, msgsOpts...)
}

// ModelMessagesSlice generates a slice of random model.Messages with the message and messages options.
func ModelMessagesSlice(
	n uint,
	msgCount uint,
	msgOpts []ModelMessageOption,
	msgsOpts ...ModelMessagesOption,
) []*model.Messages {
	// msgsF wraps ModelMessages() for slice function.
	msgsF := func(opts ...ModelMessagesOption) *model.Messages {
		return ModelMessages(msgCount, msgOpts, opts...)
	}

	return slice(n, msgsF, msgsOpts...)
}

// ------------------Messages Opts------------------

type ModelMessagesOption func(*model.Messages)

// WithLanguage sets the language of the model.Messages.
func WithLanguage(lang language.Tag) ModelMessagesOption {
	return func(m *model.Messages) {
		m.Language = lang
	}
}
