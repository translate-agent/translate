package server

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

const mockTranslation = "{Translated}"

type mockTranslator struct{}

func (m *mockTranslator) Translate(ctx context.Context, messages *model.Translation, targetLanguage language.Tag) (*model.Translation, error) {
	newMessages := &model.Translation{
		Language: targetLanguage,
		Messages: make([]model.Message, 0, len(messages.Messages)),
		Original: messages.Original,
	}

	newMessages.Messages = append(newMessages.Messages, messages.Messages...)

	for i := range newMessages.Messages {
		newMessages.Messages[i].Message = mockTranslation
		newMessages.Messages[i].Status = model.MessageStatusFuzzy
	}

	return newMessages, nil
}

func Test_fuzzyTranslate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	translateSrv := NewTranslateServiceServer(nil, &mockTranslator{})

	originalMessages1 := randOriginalMessages(3)
	originalMessages2 := randOriginalMessages(10)

	tests := []struct {
		name               string
		originalMessages   *model.Translation
		translatedMessages []model.Translation
		assertFunc         func(t *testing.T, originalMessages *model.Translation, translatedMessages []model.Translation)
	}{
		{
			name:               "Fuzzy translate untranslated translation for one translation",
			originalMessages:   originalMessages1,
			translatedMessages: randTranslatedMessages(1, 3, originalMessages1),
		},
		{
			name:               "Fuzzy translate untranslated translation for five translations",
			originalMessages:   originalMessages2,
			translatedMessages: randTranslatedMessages(5, 5, originalMessages2),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			allMessages := make(model.TranslationSlice, 0, len(tt.translatedMessages)+1)
			allMessages = append(allMessages, *tt.originalMessages)
			allMessages = append(allMessages, tt.translatedMessages...)

			untranslatedMessageIDLookup := make(map[string]struct{})

			// Randomly set message status to untranslated
			for _, msg := range tt.originalMessages.Messages {
				if gofakeit.Bool() {
					untranslatedMessageIDLookup[msg.ID] = struct{}{}
				}
			}

			for i := range allMessages {
				if allMessages[i].Original {
					continue
				}

				for j := range allMessages[i].Messages {
					if _, ok := untranslatedMessageIDLookup[allMessages[i].Messages[j].ID]; ok {
						allMessages[i].Messages[j].Status = model.MessageStatusUntranslated
					}
				}
			}

			err := translateSrv.fuzzyTranslate(ctx, allMessages)
			require.NoError(t, err)

			// Check that untranslated translation have been translated and marked as fuzzy for all translations.
			for _, m := range allMessages {
				if m.Original {
					require.Equal(t, *tt.originalMessages, m)
					continue
				}

				for _, message := range m.Messages {
					if _, ok := untranslatedMessageIDLookup[message.ID]; ok {
						require.Equal(t, mockTranslation, message.Message)
						require.Equal(t, model.MessageStatusFuzzy.String(), message.Status.String())
					} else {
						require.Equal(t, model.MessageStatusTranslated.String(), message.Status.String())
					}
				}
			}
		})
	}
}

// helpers

// randOriginalMessages creates a random translation with the original flag set to true.
func randOriginalMessages(messageCount uint) *model.Translation {
	return rand.ModelTranslation(
		messageCount,
		[]rand.ModelMessageOption{rand.WithStatus(model.MessageStatusTranslated)},
		rand.WithOriginal(true),
		rand.WithLanguage(language.English))
}

// randTranslatedMessages creates a random translation with the original flag set to false
// with the same IDs as the original translation, and with translated status.
func randTranslatedMessages(n uint, msgCount uint, original *model.Translation) []model.Translation {
	messages := make([]model.Translation, n)
	for i, lang := range rand.Languages(n) {
		messages[i] = *rand.ModelTranslation(
			msgCount,
			[]rand.ModelMessageOption{rand.WithStatus(model.MessageStatusTranslated)},
			rand.WithOriginal(false),
			rand.WithSameIDs(original),
			rand.WithLanguage(lang))
	}

	return messages
}
