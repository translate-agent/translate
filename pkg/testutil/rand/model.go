package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type modelType interface {
	model.Service | model.Message | model.Translation
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

// ModelServices generates a slice of random model.Service using the provided options for each service.
func ModelServices(n uint, opts ...ModelServiceOption) []*model.Service {
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
	msg := &model.Message{
		ID:          gofakeit.SentenceSimple(),
		PluralID:    gofakeit.SentenceSimple(),
		Message:     gofakeit.SentenceSimple(),
		Description: gofakeit.SentenceSimple(),
		Status:      MessageStatus(),
	}

	if gofakeit.Bool() {
		gofakeit.Slice(&msg.Positions)
	}

	return msg
}

// ModelMessage generates a random model.Message using provided options.
func ModelMessage(opts ...ModelMessageOption) *model.Message {
	opts = append(opts, WithPluralID("")) // PluralID is ignored as it is not used in the repo yet

	return mdl(modelMessage, opts...)
}

// ModelMessages generates a slice of random model.Message using the provided options for each message.
func ModelMessages(n uint, opts ...ModelMessageOption) []*model.Message {
	return slice(n, ModelMessage, opts...)
}

// MessageStatus returns a random model.MessageStatus.
func MessageStatus() model.MessageStatus {
	return model.MessageStatus(gofakeit.IntRange(0, 2)) //nolint:gomnd
}

// ------------------Message Opts------------------

type ModelMessageOption func(*model.Message)

// WithPluralID sets the PluralID of the model.Message to "".
func WithPluralID(pluralID string) ModelMessageOption {
	return func(m *model.Message) {
		m.PluralID = pluralID
	}
}

func WithMessage(msg string) ModelMessageOption {
	return func(m *model.Message) {
		m.Message = msg
	}
}

func WithStatus(status model.MessageStatus) ModelMessageOption {
	return func(m *model.Message) {
		m.Status = status
	}
}

// WithMessageFormat encloses the message in curly braces.
func WithMessageFormat() ModelMessageOption {
	return func(m *model.Message) {
		m.Message = "{" + m.Message + "}"
	}
}

// ------------------Translation-----------------

// modelTranslation generates a random model.Translation with the given
// count of messages and using the provided options for each message.
func modelTranslation(msgCount uint, msgOpts ...ModelMessageOption) *model.Translation {
	translation := &model.Translation{
		Language: Language(),
		Messages: make([]model.Message, msgCount),
	}

	if msgCount == 0 {
		return translation
	}

	msgs := ModelMessages(msgCount, msgOpts...)
	for i, msg := range msgs {
		translation.Messages[i] = *msg
	}

	return translation
}

// ModelTranslation generates a random model.Translation
// with specific message count, message and translation options.
func ModelTranslation(msgCount uint,
	msgOpts []ModelMessageOption,
	translationOpts ...ModelTranslationOption,
) *model.Translation {
	// translationF wraps modelTranslation() for mdl function.
	translationF := func() *model.Translation {
		return modelTranslation(msgCount, msgOpts...)
	}

	return mdl(translationF, translationOpts...)
}

// ModelTranslations generates a slice of random model.Translation with the message and translation options.
func ModelTranslations(
	n uint,
	msgCount uint,
	msgOpts []ModelMessageOption,
	translationOpts ...ModelTranslationOption,
) []*model.Translation {
	// translationF wraps ModelTranslation() for slice function.
	translationF := func(opts ...ModelTranslationOption) *model.Translation {
		return ModelTranslation(msgCount, msgOpts, opts...)
	}

	return slice(n, translationF, translationOpts...)
}

// ------------------Translation Opts------------------

type ModelTranslationOption func(*model.Translation)

// WithLanguage sets the language of the model.Translation.
func WithLanguage(lang language.Tag) ModelTranslationOption {
	return func(t *model.Translation) {
		t.Language = lang
	}
}

// WithOriginal sets the original flag of the model.Translation.
func WithOriginal(original bool) ModelTranslationOption {
	return func(t *model.Translation) {
		t.Original = original
	}
}

func WithSameIDs(t *model.Translation) ModelTranslationOption {
	return func(t2 *model.Translation) {
		for i := range t2.Messages {
			t2.Messages[i].ID = t.Messages[i].ID
		}
	}
}
