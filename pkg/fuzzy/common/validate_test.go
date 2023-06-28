package common

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/fuzzy/translatetest"
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
			messages:    translatetest.RandMessages(5, language.English),
			targetLang:  language.Latvian,
			expectedErr: nil,
		},
		{
			name:        "Nil messages",
			messages:    nil,
			targetLang:  language.German,
			expectedErr: ErrNilMessages,
		},
		{
			name:        "No messages",
			messages:    translatetest.RandMessages(0, language.English),
			targetLang:  language.Latvian,
			expectedErr: ErrNoMessages,
		},
		{
			name:        "Undefined target language",
			messages:    translatetest.RandMessages(5, language.English),
			targetLang:  language.Und,
			expectedErr: ErrTargetLangUndefined,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateTranslate(tt.messages, tt.targetLang)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
