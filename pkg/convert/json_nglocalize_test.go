package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromNG_JSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		data    []byte
		name    string
		want    model.Messages
	}{
		{
			name: "All OK",
			data: []byte(`
      {
        "locale": "fr",
        "translations": {
          "Hello": "Bonjour",
          "Welcome": "Bienvenue"
        }
      }
`),
			want: model.Messages{
				Language: language.French,
				Messages: []model.Message{
					{
						ID:      "Hello",
						Message: "Bonjour",
					},
					{
						ID:      "Welcome",
						Message: "Bienvenue",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			data: []byte(`
      {
        "locale": "xyz-ZY-Latn",
        "translations": {
          "Hello": "Bonjour",
          "Welcome": "Bienvenue"
        }
      }
`),
			wantErr: fmt.Errorf("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messages, err := FromNgLocalize(tt.data)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.want.Language, messages.Language)
			assert.ElementsMatch(t, tt.want.Messages, messages.Messages)
		})
	}
}

func TestToNG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		want     []byte
		wantErr  error
		messages model.Messages
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
      }
`),
			messages: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Welcome",
						Message:     "Welcome to our website!",
						Description: "To welcome a new visitor",
					},
					{
						ID:          "Error",
						Message:     "Something went wrong. Please try again later.",
						Description: "To inform the user of an error",
					},
					{
						ID:      "Feedback",
						Message: "We appreciate your feedback. Thank you for using our service.",
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToNgLocalize(tt.messages)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.JSONEq(t, string(tt.want), string(result))
		})
	}
}
