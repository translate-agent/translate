package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_ValidateTranslate(t *testing.T) {
	t.Parallel()

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
			expectedErr: errNilMessages,
		},
		{
			name:        "No messages",
			messages:    randMessages(0, language.English),
			targetLang:  language.Latvian,
			expectedErr: errNoMessages,
		},
		{
			name:        "Undefined target language",
			messages:    randMessages(5, language.English),
			targetLang:  language.Und,
			expectedErr: errTargetLangUndefined,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateTranslate(tt.messages, tt.targetLang)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
