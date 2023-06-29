package fuzzy

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/translate"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// ---–––--------------Actual Tests------------------–––---

func Test_TranslateMock(t *testing.T) {
	t.Parallel()

	allMocks(t, func(t *testing.T, mock Translator) {
		tests := []struct {
			targetLang  language.Tag
			expectedErr error
			messages    *model.Messages
			name        string
		}{
			{
				name:       "One message",
				messages:   randMessages(1, language.English),
				targetLang: language.Latvian,
			},
			{
				name:       "Multiple messages",
				messages:   randMessages(5, language.Latvian),
				targetLang: language.German,
			},
			{
				name:       "Undefined messages language",
				messages:   randMessages(5, language.Und),
				targetLang: language.English,
			},
			{
				name:        "Unsupported target language",
				messages:    randMessages(5, language.German),
				targetLang:  language.Afrikaans,
				expectedErr: errors.New("unsupported language"),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				translatedMsgs, err := mock.Translate(context.Background(), tt.messages, tt.targetLang)

				if tt.expectedErr != nil {
					require.ErrorContains(t, err, tt.expectedErr.Error())
					return
				}

				require.NoError(t, err)

				// Check the that the translated messages have the correct language.
				require.Equal(t, tt.targetLang, translatedMsgs.Language)

				// Check that length matches.
				require.Len(t, translatedMsgs.Messages, len(tt.messages.Messages))

				for i, m := range translatedMsgs.Messages {
					// Check the translated messages are not empty and are marked as fuzzy.
					require.NotEmpty(t, m.Message)
					require.True(t, m.Fuzzy)

					// Reset the message to empty and fuzzy to original values, for the last check for side effects.
					translatedMsgs.Messages[i].Message = tt.messages.Messages[i].Message
					translatedMsgs.Messages[i].Fuzzy = tt.messages.Messages[i].Fuzzy
				}

				// Check the translated messages are the same as the input messages. (Check for side effects)
				require.ElementsMatch(t, tt.messages.Messages, translatedMsgs.Messages)
			})
		}
	})
}

// -------------------------Mocks------------------------------

// ---–––--------------Google Translate------------------–––---

// MockGoogleTranslateClient is a mock implementation of the Google Translate client.
type MockGoogleTranslateClient struct{}

// Translate mocks the Translate method of the Google Translate client.
func (m *MockGoogleTranslateClient) Translate(
	_ context.Context,
	inputs []string,
	target language.Tag,
	_ *translate.Options,
) ([]translate.Translation, error) {
	// Mock the Bad request error for unsupported language.Afrikaans.
	if target == language.Afrikaans {
		return nil, errors.New("mock: bad request: unsupported language")
	}

	translations := make([]translate.Translation, 0, len(inputs))

	for range inputs {
		translations = append(translations, translate.Translation{Text: gofakeit.SentenceSimple()})
	}

	return translations, nil
}

func (m *MockGoogleTranslateClient) Close() error { return nil }

// -----------------------Helpers and init----------------------------

var mockTranslators map[string]Translator

func init() {
	mockTranslators = make(map[string]Translator, len(SupportedServices))

	// Google Translate
	gt, _, _ := NewGoogleTranslate(
		context.Background(),
		WithClient(&MockGoogleTranslateClient{}),
	)

	mockTranslators["GoogleTranslate"] = gt
}

// allMocks runs a test function f for each mocked translate service that is defined in the mockTranslators map.
func allMocks(t *testing.T, f func(t *testing.T, mock Translator)) {
	for name, mock := range mockTranslators {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f(t, mock)
		})
	}
}

// randMessages returns a random messages model with the given count of messages and source language.
// The messages will not be fuzzy.
func randMessages(msgCount uint, srcLang language.Tag) *model.Messages {
	msgOpts := []rand.ModelMessageOption{rand.WithFuzzy(false)}
	msgsOpts := []rand.ModelMessagesOption{rand.WithLanguage(srcLang)}

	return rand.ModelMessages(msgCount, msgOpts, msgsOpts...)
}
