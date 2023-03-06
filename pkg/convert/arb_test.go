package convert

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want    model.Messages
		wantErr error
		name    string
		data    []byte
	}{
		{
			name: "Combination of messages",
			data: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": "Message to greet the World"
				},
				"greeting": "Welcome {user}!",
				"@greeting": {
					"placeholders": {
						"user": {
							"type": "string",
							"example": "Bob"
						}
					}
				},
				"farewell": "Goodbye friend"
			}			
				`),
			want: model.Messages{
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "Hello World!",
						Description: "Message to greet the World",
					},
					{
						ID:      "greeting",
						Message: "Welcome {user}!",
					},
					{
						ID:      "farewell",
						Message: "Goodbye friend",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Wrong value type for @title",
			data: []byte(`
			{
				"title": "Hello World!",
				"@title": "Message to greet the World"
			}			
					`),
			wantErr: errors.New("expected a map, got 'string'"),
		},
		{
			name: "Wrong value type for greeting key",
			data: []byte(`
			{
				"title": "Hello World!",
				"greeting": {
					"description": "Needed for greeting"
				}
			}			
					`),
			wantErr: errors.New("unsupported value type 'map[string]interface {}' for key 'greeting'"),
		},
		{
			name: "Wrong value type for description key",
			data: []byte(`
			{
				"title": "Hello World!",
				"@title": {
					"description": {
						"meaning": "When you greet someone"
					}
				}
			}			
					`),
			wantErr: errors.New("'Description' expected type 'string', got unconvertible type 'map[string]interface {}'"),
		},
		{
			name: "With locale",
			data: []byte(`
      {
        "@@locale": "en",
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }
			`),
			want: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "Hello World!",
						Description: "Message to greet the World",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "With malformed locale",
			data: []byte(`
      {
        "@@locale": "asd-gh-jk",
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }
			`),
			wantErr: fmt.Errorf("language: tag is not well-formed"),
		},
		{
			name: "With wrong value type for locale",
			data: []byte(`
      {
        "@@locale": {
          "tag": "fr-FR"
        },
        "title": "Hello World!",
        "@title": {
          "description": "Message to greet the World"
        }
      }

			`),
			wantErr: fmt.Errorf("unsupported value type 'map[string]interface {}' for key '@@locale'"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := FromArb(tt.data)
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.want.Language, res.Language)
			assert.ElementsMatch(t, tt.want.Messages, res.Messages)
		})
	}
}

func Test_ToArb(t *testing.T) {
	t.Parallel()

	messages := model.Messages{
		Language: language.French,
		Messages: []model.Message{
			{
				ID:          "title",
				Message:     "Hello World!",
				Description: "Message to greet the World",
			},
			{
				ID:      "greeting",
				Message: "Welcome {user}",
			},
		},
	}

	want := []byte(`
	{
		"@@locale":"fr",
		"title":"Hello World!",
		"@title":{
			"description":"Message to greet the World"
		},
		"greeting":"Welcome {user}"
	}
	`)

	res, err := ToArb(messages)
	if !assert.NoError(t, err) {
		return
	}

	assert.JSONEq(t, string(want), string(res))
}
