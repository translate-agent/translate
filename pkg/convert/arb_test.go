package convert

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr string
		name    string
		input   []byte
		want    model.Translation
	}{
		// Positive tests
		{
			name: "Message with special chars",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": "Message to greet the World"
				},
				"greeting": "Welcome {user} | \\ !",
				"@greeting": {
					"placeholders": {
						"user": {
							"type": "string",
							"example": "Bob"
						}
					}
				},
				"farewell": "Goodbye friend"
			}`),
			want: model.Translation{
				Original: true,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "Hello World!",
						Description: `Message to greet the World`,
						Status:      model.MessageStatusTranslated,
					},
					{
						ID:      "greeting",
						Message: `Welcome \{user\} | \\ !`,
						Status:  model.MessageStatusTranslated,
					},
					{
						ID:      "farewell",
						Message: `Goodbye friend`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "With locale",
			input: []byte(`
      {
        "@@locale": "lv",
        "title": "",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			want: model.Translation{
				Language: language.Latvian,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     ``,
						Description: "Message to greet the World",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		// Negative tests
		{
			name: "Wrong value type for @title",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": "Message to greet the World"
			}`),
			wantErr: `find description of "title": decode metadata map: '' expected a map, got 'string'`,
		},
		{
			name: "Wrong value type for greeting key",
			input: []byte(`
			{
				"title": "Hello World!",
				"greeting": {
					"description": "Needed for greeting"
				}
			}`),
			wantErr: "unsupported value type 'map[string]interface {}' for key 'greeting'",
		},
		{
			name: "Wrong value type for description key",
			input: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": {
						"meaning": "When you greet someone"
					}
				}
			}`),
			wantErr: `find description of "title": decode metadata map: 1 error(s) decoding:

* 'description' expected type 'string', got unconvertible type 'map[string]interface {}', value: 'map[meaning:When you greet someone]'`, //nolint:lll
		},
		{
			name: "With malformed locale",
			input: []byte(`
      {
        "@@locale": "asd-gh-jk",
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			wantErr: "find locale: parse language: language: tag is not well-formed", //nolint:dupword
		},
		{
			name: "With wrong value type for locale",
			input: []byte(`
      {
        "@@locale": {
          "tag": "fr-FR"
        },
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }`),
			wantErr: `find locale: unsupported value type "map[string]interface {}" for key "@@locale"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromArb(test.input, &test.want.Original)
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

func Test_ToArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		want  []byte
		input model.Translation
	}{
		{
			name: "valid input",
			input: model.Translation{
				Language: language.French,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     `Hello World!`,
						Description: "Message to greet the World",
					},
					{
						ID:      "greeting",
						Message: `Welcome Sion`,
					},
				},
			},
			want: []byte(`
	{
		"@@locale":"fr",
		"title":"Hello World!",
		"@title":{
			"description":"Message to greet the World"
		},
		"greeting":"Welcome Sion"
	}`),
		},
		{
			name: "Message with special chars",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "title",
						Message: `Hello World!`,
					},
					{
						ID:      "greeting",
						Message: `Welcome \{user\} | \\ !`,
					},
					{
						ID:      "farewell",
						Message: `Goodbye friend`,
					},
				},
			},
			want: []byte(`
			{
				"@@locale":"en",
				"farewell":"Goodbye friend",
				"greeting":"Welcome {user} | \\ !",
				"title":"Hello World!"
			}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToArb(test.input)
			if err != nil {
				t.Error(err)
				return
			}

			var want, got any

			if err = json.Unmarshal(actual, &got); err != nil {
				t.Error(err)
				return
			}

			if err = json.Unmarshal(test.want, &want); err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("want %s, got %s", test.want, actual)
			}
		})
	}
}
