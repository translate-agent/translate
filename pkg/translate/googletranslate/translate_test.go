package googletranslate

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/translate"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/translate/translatetest"
	"golang.org/x/text/language"
)

// mockGoogleTranslateClient is a mock implementation of the Google Translate client.
type mockGoogleTranslateClient struct{}

// Translate mocks the Translate method of the Google Translate client.
func (m *mockGoogleTranslateClient) Translate(
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

func (m *mockGoogleTranslateClient) Close() error { return nil }

// Test_Translate tests the Translate method of the Google Translate service using a mock client.
func Test_Translate(t *testing.T) {
	t.Parallel()

	mockClient := &mockGoogleTranslateClient{}

	mockGoogleTranslate, _, _ := NewGoogleTranslate(context.Background(), WithClient(mockClient))

	tests := []struct {
		targetLang  language.Tag
		expectedErr error
		messages    *model.Messages
		name        string
	}{
		{
			name:       "One message",
			messages:   translatetest.RandMessages(1, language.English),
			targetLang: language.Latvian,
		},
		{
			name:       "Multiple messages",
			messages:   translatetest.RandMessages(5, language.Latvian),
			targetLang: language.German,
		},
		{
			name:       "Undefined messages language",
			messages:   translatetest.RandMessages(5, language.Und),
			targetLang: language.English,
		},
		{
			name:        "Unsupported target language",
			messages:    translatetest.RandMessages(5, language.German),
			targetLang:  language.Afrikaans,
			expectedErr: errors.New("unsupported language"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			translatedMsgs, err := mockGoogleTranslate.Translate(context.Background(), tt.messages, tt.targetLang)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			// Check the that the translated messages have the correct language.
			require.Equal(t, tt.targetLang, translatedMsgs.Language)

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
}
