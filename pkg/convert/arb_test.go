package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
)

func Test_FromArb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		want    model.Messages
		wantErr bool
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
			wantErr: false,
		},
		{
			name: "Wrong value type for @key",
			data: []byte(`
			{
				"title": "Hello World!",
				"@title": "Message to greet the World"
			}			
					`),
			wantErr: true,
		},
		{
			name: "Wrong value type for standard key",
			data: []byte(`
			{
				"title": "Hello World!",
				"greeting": {
					"description": "Needed for greeting"
				}
			}			
					`),
			wantErr: true,
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
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := FromArb(tt.data)

			if tt.wantErr {
				assert.Error(t, err)
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
