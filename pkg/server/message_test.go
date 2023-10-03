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

func (m *mockTranslator) Translate(ctx context.Context, translation *model.Translation, targetLanguage language.Tag) (*model.Translation, error) {
	newTranslation := &model.Translation{
		Language: targetLanguage,
		Messages: make([]model.Message, 0, len(translation.Messages)),
		Original: translation.Original,
	}

	newTranslation.Messages = append(newTranslation.Messages, translation.Messages...)

	for i := range newTranslation.Messages {
		newTranslation.Messages[i].Message = mockTranslation
		newTranslation.Messages[i].Status = model.MessageStatusFuzzy
	}

	return newTranslation, nil
}

func Test_fuzzyTranslate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	translateSrv := NewTranslateServiceServer(nil, &mockTranslator{})

	originalTranslation1 := randOriginalTranslation(3)
	originalTranslation2 := randOriginalTranslation(10)

	tests := []struct {
		name               string
		originalTranslation   *model.Translation
		translatedTranslation []model.Translation
		assertFunc         func(t *testing.T, originalTranslation *model.Translation, translatedTranlation []model.Translation)
	}{
		{
			name:               "Fuzzy translate untranslated translation for one message",
			originalTranslation:   originalTranslation1,
			translatedTranslation: randTranslatedTranslation(1, 3, originalTranslation1),
		},
		{
			name:               "Fuzzy translate untranslated translation for five messages",
			originalTranslation:   originalTranslation2,
			translatedTranslation: randTranslatedTranslation(5, 5, originalTranslation2),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			allTranslations := make(model.Translations, 0, len(tt.translatedTranslation)+1)
			allTranslations = append(allTranslations, *tt.originalTranslation)
			allTranslations = append(allTranslations, tt.translatedTranslation...)

			untranslatedMessageIDLookup := make(map[string]struct{})

			// Randomly set message status to untranslated
			for _, msg := range tt.originalTranslation.Messages {
				if gofakeit.Bool() {
					untranslatedMessageIDLookup[msg.ID] = struct{}{}
				}
			}

			for i := range allTranslations {
				if allTranslations[i].Original {
					continue
				}

				for j := range allTranslations[i].Messages {
					if _, ok := untranslatedMessageIDLookup[allTranslations[i].Messages[j].ID]; ok {
						allTranslations[i].Messages[j].Status = model.MessageStatusUntranslated
					}
				}
			}

			err := translateSrv.fuzzyTranslate(ctx, allTranslations)
			require.NoError(t, err)

			// Check that untranslated translation have been translated and marked as fuzzy for all translations.
			for _, translation := range allTranslations {
				if translation.Original {
					require.Equal(t, *tt.originalTranslation, translation)
					continue
				}

				for _, message := range translation.Messages {
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

// randOriginalTranslation creates a random translation with the original flag set to true.
func randOriginalTranslation(messageCount uint) *model.Translation {
	return rand.ModelTranslation(
		messageCount,
		[]rand.ModelMessageOption{rand.WithStatus(model.MessageStatusTranslated)},
		rand.WithOriginal(true),
		rand.WithLanguage(language.English))
}

// randTranslatedTranslation creates a random translation with the original flag set to false
// with the same IDs as the original translation, and with translated status.
func randTranslatedTranslation(n uint, msgCount uint, original *model.Translation) []model.Translation {
	translations := make([]model.Translation, n)
	for i, lang := range rand.Languages(n) {
		translations[i] = *rand.ModelTranslation(
			msgCount,
			[]rand.ModelMessageOption{rand.WithStatus(model.MessageStatusTranslated)},
			rand.WithOriginal(false),
			rand.WithSameIDs(original),
			rand.WithLanguage(lang))
	}

	return translations
}
