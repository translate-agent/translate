package convert

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/expect"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromNgLocalize(tt.input, &tt.want.Original)

			if tt.wanterr != nil {
				expect.ErrorContains(t, err, tt.wanterr.Error())
				return
			}

			expect.NoError(t, err)

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("\nwant %v\ngot  %v", tt.want, got)
			}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToNgLocalize(tt.input)

			if tt.wantErr != nil {
				expect.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			expect.NoError(t, err)

			assert.JSONEq(t, string(tt.want), string(got))
		})
	}
}
