package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

func Test_FromNgLocalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr error
		input       []byte
		name        string
		expected    model.Messages
	}{
		// Positive tests
		{
			name: "Original",
			input: []byte(`
      {
        "locale": "fr",
        "translations": {
          "Hello": "Bonjour",
          "Welcome": "Bienvenue"
        }
      }`),
			expected: model.Messages{
				Language: language.French,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "Hello",
						Message: "{Bonjour}",
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "Welcome",
						Message: "{Bienvenue}",
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "Translated",
			input: []byte(`
			{
				"locale": "fr",
				"translations": {
					"Hello": "Bonjour",
					"Welcome": ""
				}
			}
			`),
			expected: model.Messages{
				Language: language.French,
				Original: false,
				Messages: []model.Message{
					{
						ID:      "Hello",
						Message: "{Bonjour}",
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "Welcome",
						Message: "",
						Status:  model.MessageStatusUntranslated,
					},
				},
			},
		},
		// Negative tests
		{
			name: "Malformed language",
			input: []byte(`
      {
        "locale": "xyz-ZY-Latn",
        "translations": {
          "Hello": "Bonjour",
          "Welcome": "Bienvenue"
        }
      }`),
			expectedErr: fmt.Errorf("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromNgLocalize(tt.input, tt.expected.Original)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			testutil.EqualMessages(t, &tt.expected, &actual)
		})
	}
}

func Test_ToNgLocalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expected    []byte
		expectedErr error
		input       model.Messages
	}{
		{
			name: "All OK",
			expected: []byte(`
      {
        "locale": "en",
        "translations": {
          "Welcome": "Welcome to our website!",
          "Error": "Something went wrong. Please try again later.",
          "Feedback": "We appreciate your feedback. Thank you for using our service."
        }
      }`),
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Welcome",
						Message:     "{Welcome to our website!}",
						Description: "To welcome a new visitor",
					},
					{
						ID:          "Error",
						Message:     "{Something went wrong. Please try again later.}",
						Description: "To inform the user of an error",
					},
					{
						ID:      "Feedback",
						Message: "{We appreciate your feedback. Thank you for using our service.}",
					},
				},
			},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToNgLocalize(tt.input)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.JSONEq(t, string(tt.expected), string(actual))
		})
	}
}
