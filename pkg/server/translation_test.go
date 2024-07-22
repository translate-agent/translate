package server

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/expect"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

const mockTranslation = "{Translated}"

type mockTranslator struct{}

func (m *mockTranslator) Translate(ctx context.Context,
	translation *model.Translation,
	targetLanguage language.Tag,
) (*model.Translation, error) {
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

	translateSrv := NewTranslateServiceServer(nil, &mockTranslator{})
	originalTranslation1 := randOriginalTranslation(3)
	originalTranslation2 := randOriginalTranslation(10)

	tests := []struct {
		name                string
		originalTranslation *model.Translation
		translations        []model.Translation
	}{
		{
			name:                "Fuzzy translate untranslated messages for one translation",
			originalTranslation: originalTranslation1,
			translations:        randTranslations(1, 3, originalTranslation1),
		},
		{
			name:                "Fuzzy translate untranslated messages for five translations",
			originalTranslation: originalTranslation2,
			translations:        randTranslations(5, 5, originalTranslation2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			allTranslations := append(model.Translations{*tt.originalTranslation}, tt.translations...)
			untranslatedMessageIDLookup := randomUntranslatedMessageStatus(t, allTranslations)

			expect.NoError(t, translateSrv.fuzzyTranslate(context.Background(), allTranslations))

			// Check that untranslated messages have been translated and marked as fuzzy for all translations.
			for _, translation := range allTranslations {
				if translation.Original {
					require.Equal(t, *tt.originalTranslation, translation)
					continue
				}

				for _, message := range translation.Messages {
					if _, ok := untranslatedMessageIDLookup[message.ID]; ok {
						expect.Equal(t, mockTranslation, message.Message)
						expect.Equal(t, model.MessageStatusFuzzy.String(), message.Status.String())
					} else {
						expect.Equal(t, model.MessageStatusTranslated.String(), message.Status.String())
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

// randTranslations creates random translations with the original flag set to false
// with the same IDs as the original translation, and with translated status.
func randTranslations(n uint, msgCount uint, original *model.Translation) []model.Translation {
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

// randomUntranslatedMessageStatus randomly sets the message status to untranslated,
// starting with the original and then remaining translations.
// Returns a map containing the untranslated message IDs.
func randomUntranslatedMessageStatus(t *testing.T, translations model.Translations) map[string]struct{} {
	t.Helper()

	origIdx := translations.OriginalIndex()

	if origIdx == -1 {
		return nil
	}

	untranslatedMessageIDLookup := make(map[string]struct{})

	for _, v := range translations[origIdx].Messages {
		untranslatedMessageIDLookup[v.ID] = struct{}{}
	}

	// Randomly set message status to untranslated
	for _, msg := range translations[origIdx].Messages {
		if gofakeit.Bool() {
			untranslatedMessageIDLookup[msg.ID] = struct{}{}
		}
	}

	for i := range translations {
		if translations[i].Original {
			continue
		}

		for j := range translations[i].Messages {
			if _, ok := untranslatedMessageIDLookup[translations[i].Messages[j].ID]; ok {
				translations[i].Messages[j].Status = model.MessageStatusUntranslated
			}
		}
	}

	return untranslatedMessageIDLookup
}
