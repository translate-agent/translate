package convert

import (
	"errors"
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
		wanterr error
		input   []byte
		name    string
		want    model.Translation
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
			want: model.Translation{
				Language: language.French,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "Hello",
						Message: `Bonjour`,
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "Welcome",
						Message: `Bienvenue`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "Message with special chars",
			input: []byte(`
      {
        "locale": "fr",
        "translations": {
          "Hello": "Welcome {user}! | \\",
          "Welcome": "Bienvenue"
        }
      }`),
			want: model.Translation{
				Language: language.French,
				Original: true,
				Messages: []model.Message{
					{
						ID:      "Hello",
						Message: `Welcome \{user\}! | \\`,
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "Welcome",
						Message: `Bienvenue`,
						Status:  model.MessageStatusTranslated,
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
			wanterr: errors.New("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromNgLocalize(test.input, &test.want.Original)

			if test.wanterr != nil {
				require.ErrorContains(t, err, test.wanterr.Error())
				return
			}

			require.NoError(t, err)

			testutil.EqualTranslations(t, &test.want, &got)
		})
	}
}

func Test_ToNgLocalize(t *testing.T) {
	t.Parallel()

	t.Skip() // TODO

	tests := []struct {
		name    string
		want    []byte
		wantErr error
		input   model.Translation
	}{
		{
			name: "All OK",
			want: []byte(`
      {
        "locale": "en",
        "translations": {
          "Welcome": "Welcome to our website!",
          "Error": "Something went wrong. Please try again later.",
          "Feedback": "We appreciate your feedback. Thank you for using our service."
        }
      }`),
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Welcome",
						Message:     `Welcome to our website!`,
						Description: "To welcome a new visitor",
					},
					{
						ID:          "Error",
						Message:     `Something went wrong. Please try again later.`,
						Description: "To inform the user of an error",
					},
					{
						ID:      "Feedback",
						Message: `We appreciate your feedback. Thank you for using our service.`,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Message with special chars",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Welcome",
						Message:     `Welcome to our website \{user\} 99|100 \\`,
						Description: "To welcome a new visitor",
					},
				},
			},
			want:    []byte(`{"locale": "en","translations": {"Welcome": "Welcome to our website {user} 99|100 \\"}}`),
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToNgLocalize(test.input)

			if test.wantErr != nil {
				require.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			require.NoError(t, err)

			assert.JSONEq(t, string(test.want), string(got))
		})
	}
}
