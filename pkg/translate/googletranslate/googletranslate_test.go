package googletranslate

import (
	"context"
	"errors"
	"testing"

	googleTranslate "cloud.google.com/go/translate"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// mockGoogleClient is a mock implementation of the Google Translate client.
type mockGoogleClient struct{}

// mockGoogleClient.Translate mocks the Translate method of the Google Translate client.
func (m *mockGoogleClient) Translate(
	_ context.Context,
	inputs []string,
	_ language.Tag,
	_ *googleTranslate.Options,
) ([]googleTranslate.Translation, error) {
	translations := make([]googleTranslate.Translation, 0, len(inputs))
	for range inputs {
		translations = append(translations, googleTranslate.Translation{
			Text:   gofakeit.SentenceSimple(),
			Source: language.English,
		})
	}

	return translations, nil
}

// mockGoogleClient.SupportedLanguages mocks the SupportedLanguages method of the Google Translate client.
func (m *mockGoogleClient) SupportedLanguages(context.Context, language.Tag) ([]googleTranslate.Language, error) {
	fakeSupportedLangs := []googleTranslate.Language{
		{
			Tag: language.English,
		},
		{
			Tag: language.Latvian,
		},
		{
			Tag: language.German,
		},
	}

	return fakeSupportedLangs, nil
}

func (m *mockGoogleClient) Close() error { return nil }

func randMessages(count uint, lang language.Tag) *model.Messages {
	opts := []rand.ModelMessagesOption{rand.WithoutTranslations(), rand.WithLanguage(lang)}
	return rand.ModelMessages(count, opts...)
}

func Test_ValidateTranslateReq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockGoogleTranslate, err := NewGoogleTranslate(ctx, &mockGoogleClient{})
	require.NoError(t, err)

	defer mockGoogleTranslate.client.Close()

	tests := []struct {
		targetLang  language.Tag
		expectedErr error
		messages    *model.Messages
		name        string
	}{
		{
			name:        "Valid input",
			messages:    randMessages(5, language.English),
			targetLang:  language.Latvian,
			expectedErr: nil,
		},
		{
			name:        "Nil messages",
			messages:    nil,
			targetLang:  language.German,
			expectedErr: errors.New("nil messages"),
		},
		{
			name:        "No messages",
			messages:    randMessages(0, language.English),
			targetLang:  language.Latvian,
			expectedErr: errors.New("no messages"),
		},
		{
			name:        "Undefined target language",
			messages:    randMessages(5, language.English),
			targetLang:  language.Und,
			expectedErr: errors.New("target language undefined"),
		},
		{
			name:        "Messages language not supported",
			messages:    randMessages(5, language.Chinese),
			targetLang:  language.English,
			expectedErr: errors.New("source language zh not supported"),
		},
		{
			name:        "Target language not supported",
			messages:    randMessages(5, language.English),
			targetLang:  language.Chinese,
			expectedErr: errors.New("target language zh not supported"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := mockGoogleTranslate.validateTranslateReq(tt.messages, tt.targetLang)
			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

// Test_Translate tests the Translate method of the Google Translate service using a mock client.
func Test_Translate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockGoogleTranslate, err := NewGoogleTranslate(ctx, &mockGoogleClient{})
	require.NoError(t, err)

	defer mockGoogleTranslate.client.Close()

	tests := []struct {
		messages   *model.Messages
		targetLang language.Tag
		name       string
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			translatedMsgs, err := mockGoogleTranslate.Translate(ctx, tt.messages, tt.targetLang)
			require.NoError(t, err)

			// Check the language is the same as the input language. (Check for side effects)
			require.Equal(t, tt.messages.Language, translatedMsgs.Language)

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
