package translatetest

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/translate"
	"go.expect.digital/translate/pkg/translate/googletranslate"
	"go.expect.digital/translate/pkg/translate/translatetest/mock"
	"golang.org/x/text/language"
)

var mockTranslators map[string]service

func init() {
	mockTranslators = make(map[string]service, len(translate.SupportedServices))

	// Google Translate
	gt, _, _ := googletranslate.NewGoogleTranslate(
		context.Background(),
		googletranslate.WithClient(
			&mock.MockGoogleTranslateClient{},
		),
	)
	mockTranslators["GoogleTranslate"] = gt
}

// allMocks runs a test function f for each mocked translate service that is defined in the mockTranslators map.
func allMocks(t *testing.T, f func(t *testing.T, mock service)) {
	for name, mock := range mockTranslators {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f(t, mock)
		})
	}
}

func Test_TranslateMock(t *testing.T) {
	t.Parallel()

	allMocks(t, func(t *testing.T, mock service) {
		tests := []struct {
			targetLang  language.Tag
			expectedErr error
			messages    *model.Messages
			name        string
		}{
			{
				name:       "One message",
				messages:   RandMessages(1, language.English),
				targetLang: language.Latvian,
			},
			{
				name:       "Multiple messages",
				messages:   RandMessages(5, language.Latvian),
				targetLang: language.German,
			},
			{
				name:       "Undefined messages language",
				messages:   RandMessages(5, language.Und),
				targetLang: language.English,
			},
			{
				name:        "Unsupported target language",
				messages:    RandMessages(5, language.German),
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
