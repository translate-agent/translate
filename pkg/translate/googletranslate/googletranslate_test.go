package googletranslate

import (
	"context"
	"errors"
	"testing"

	googleTran "cloud.google.com/go/translate"
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
	ctx context.Context,
	inputs []string,
	target language.Tag,
	opts *googleTran.Options,
) ([]googleTran.Translation, error) {
	lang := rand.Lang()

	translations := make([]googleTran.Translation, 0, len(inputs))
	for range inputs {
		translations = append(translations, googleTran.Translation{
			Text:   gofakeit.SentenceSimple(),
			Source: lang,
		})
	}

	return translations, nil
}

func (m *mockGoogleClient) Close() error { return nil }

func Test_Translate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockGoogleTranslate := NewGoogleTranslate(&mockGoogleClient{})
	defer mockGoogleTranslate.Client.Close()

	// Tests

	oneMessage := rand.Messages(1, rand.WithoutTranslation())

	multipleMessages := rand.Messages(5, rand.WithoutTranslation())

	noMessages := rand.Messages(0, rand.WithoutTranslation())

	undTargetLangMessages := rand.Messages(5, rand.WithoutTranslation())

	undLangMessages := rand.Messages(5, rand.WithoutTranslation())
	undLangMessages.Language = language.Und

	tests := []struct {
		input       *model.Messages
		targetLang  language.Tag
		expectedErr error
		name        string
	}{
		{
			name:        "One message",
			input:       oneMessage,
			targetLang:  rand.Lang(),
			expectedErr: nil,
		},
		{
			name:        "Multiple messages",
			input:       multipleMessages,
			targetLang:  rand.Lang(),
			expectedErr: nil,
		},
		{
			name:        "No messages",
			input:       noMessages,
			targetLang:  rand.Lang(),
			expectedErr: errors.New("no messages"),
		},
		{
			name:        "Undefined target language",
			input:       undTargetLangMessages,
			targetLang:  language.Und,
			expectedErr: errors.New("target language undefined"),
		},
		{
			name:        "Undefined messages language",
			input:       undLangMessages,
			targetLang:  rand.Lang(),
			expectedErr: nil,
		},
		{
			name:        "Nil input",
			input:       nil,
			expectedErr: errors.New("nil messages"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			translatedMsgs, err := mockGoogleTranslate.Translate(ctx, tt.input, tt.targetLang)
			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			// Check the language is the same as the input language. (Check for side effects)
			require.Equal(t, tt.input.Language, translatedMsgs.Language)

			// Check the translated messages are not empty and are marked as fuzzy.
			for i, m := range translatedMsgs.Messages {
				require.NotEmpty(t, m.Message)
				require.True(t, m.Fuzzy)

				// Reset the message to empty and fuzzy to false for the final comparison.
				translatedMsgs.Messages[i].Message = ""
				translatedMsgs.Messages[i].Fuzzy = false
			}

			// Check the translated messages are the same as the input messages. (Check for side effects)
			require.ElementsMatch(t, tt.input.Messages, translatedMsgs.Messages)
		})
	}
}
