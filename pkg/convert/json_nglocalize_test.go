package convert

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromNgLocalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr string
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
			wantErr: `unmarshal @angular/localize JSON into ngJSON struct: language: subtag "xyz" is well-formed but unknown`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromNgLocalize(test.input, &test.want.Original)

			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			cmp := func(a, b model.Message) int {
				switch {
				default:
					return 0
				case a.ID < b.ID:
					return -1
				case b.ID < a.ID:
					return 1
				}
			}

			slices.SortFunc(test.want.Messages, cmp)
			slices.SortFunc(got.Messages, cmp)

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("\nwant %v\ngot  %v", test.want, got)
			}
		})
	}
}

func Test_ToNgLocalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    []byte
		wantErr error
		input   model.Translation
	}{
		{
			name: "All OK",
			want: []byte(`{"locale":"en","translations":{"Error":"Something went wrong. Please try again later.","Feedback":"We appreciate your feedback. Thank you for using our service.","Welcome":"Welcome to our website!"}}`), //nolint:lll
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
			want:    []byte(`{"locale":"en","translations":{"Welcome":"Welcome to our website {user} 99|100 \\"}}`),
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToNgLocalize(test.input)
			if test.wantErr != nil {
				if err.Error() != test.wantErr.Error() {
					t.Errorf("want error '%s', got '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("want equal nglocalize\n%s", diff)
			}
		})
	}
}
