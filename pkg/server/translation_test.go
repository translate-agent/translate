package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

const mockTranslation = "{Translated}"

type mockTranslator struct{}

func (m *mockTranslator) Translate(_ context.Context,
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

//nolint:gocognit
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			allTranslations := append(model.Translations{*test.originalTranslation}, test.translations...)
			untranslatedMessageIDLookup := randomUntranslatedMessageStatus(t, allTranslations)

			err := translateSrv.fuzzyTranslate(t.Context(), allTranslations)
			if err != nil {
				t.Error(err)
				return
			}

			// Check that untranslated messages have been translated and marked as fuzzy for all translations.
			for _, translation := range allTranslations {
				if translation.Original {
					if !reflect.DeepEqual(*test.originalTranslation, translation) {
						t.Errorf("\nwant %v\ngot  %v", *test.originalTranslation, translation)
					}

					continue
				}

				for _, message := range translation.Messages {
					if _, ok := untranslatedMessageIDLookup[message.ID]; ok {
						if mockTranslation != message.Message {
							t.Errorf("want message '%s', got '%s'", mockTranslation, message.Message)
						}

						if model.MessageStatusFuzzy != message.Status {
							t.Errorf("want message status '%s', got '%s'", ptr(model.MessageStatusFuzzy), &message.Status)
						}
					} else if model.MessageStatusTranslated != message.Status {
						t.Errorf("want message status '%s', got '%s'", ptr(model.MessageStatusTranslated), &message.Status)
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
