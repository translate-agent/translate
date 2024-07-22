package convert

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/expect"
	"golang.org/x/text/language"
)

func TestToGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		want  []byte
		input model.Translation
	}{
		{
			name: "valid input",
			input: model.Translation{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "1",
						Message:     `message1`,
						Description: "description1",
						Positions:   []string{"src/config.go:10", "src/config.go:20"},
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "2",
						Message:     "message2",
						Description: "description2",
					},
				},
			},
			want: []byte(`
	{
		"language":"en",
		"messages":[
			{
				"id":"1",
				"meaning":"description1",
				"message":"",
				"translation":"message1",
				"position": "src/config.go:10",
				"fuzzy":true
			},
			{
				"id":"1",
				"meaning":"description1",
				"message":"",
				"translation":"message1",
				"position": "src/config.go:20",
				"fuzzy":true
			},
			{
				"id":"2",
				"meaning":"description2",
				"message":"",
				"translation":"message2"
			}
		]
	}`),
		},
		{
			name: "Message with special chars",
			input: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "2",
						Message:     `Order #\{Id\} has been canceled for \{ClientName\} | \\`,
						Description: "description2",
						Positions:   []string{"src/config.go:20"},
						Status:      model.MessageStatusFuzzy,
					},
				},
			},
			want: []byte(`
	{
		"language": "en",
		"messages": [
			{
				"id": "2",
				"meaning": "description2",
				"message": "",
				"translation": "Order #{Id} has been canceled for {ClientName} | \\",
				"position": "src/config.go:20",
				"fuzzy":true
			}
		]
	}
	`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToGo(tt.input)
			expect.NoError(t, err)

			assert.JSONEq(t, string(tt.want), string(got))
		})
	}
}

func TestFromGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  model.Translation
	}{
		{
			name: "Valid input",
			input: []byte(`
	{
		"language": "en",
		"messages": [
			{
				"id": "1",
				"meaning": "description1",
				"message": "message1",
				"translation": "translatedMessage1",
				"position": "src/config.go:10",
				"fuzzy":true
			},
			{
				"id": "2",
				"meaning": "description2",
				"message": "message2",
				"translation": "translatedMessage2",
				"position": "src/config.go:20",
				"fuzzy":true
			},
			{
				"id": "3",
				"meaning": "description3",
				"message": "message3",
				"translation": ""
			}
		]
	}
	`),
			want: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "1",
						Message:     `translatedMessage1`,
						Description: "description1",
						Positions:   []string{"src/config.go:10"},
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "2",
						Message:     `translatedMessage2`,
						Description: "description2",
						Positions:   []string{"src/config.go:20"},
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "3",
						Message:     ``,
						Description: "description3",
						Status:      model.MessageStatusUntranslated,
					},
				},
			},
		},
		{
			name: "Message with special chars",
			input: []byte(`
	{
		"language": "en",
		"messages": [
			{
				"id": "1",
				"meaning": "description1",
				"message": "message1",
				"translation": "translatedMessage1",
				"position": "src/config.go:10",
				"fuzzy":true
			},
			{
				"id": "2",
				"meaning": "description2",
				"message": "message2",
				"translation": "Order #{Id} has been canceled for {ClientName} | \\",
				"position": "src/config.go:20",
				"fuzzy":true
			}
		]
	}
	`),
			want: model.Translation{
				Language: language.English,
				Original: false,
				Messages: []model.Message{
					{
						ID:          "1",
						Message:     `translatedMessage1`,
						Description: "description1",
						Positions:   []string{"src/config.go:10"},
						Status:      model.MessageStatusFuzzy,
					},
					{
						ID:          "2",
						Message:     `Order #\{Id\} has been canceled for \{ClientName\} | \\`,
						Description: "description2",
						Positions:   []string{"src/config.go:20"},
						Status:      model.MessageStatusFuzzy,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromGo(tt.input, &tt.want.Original)

			expect.NoError(t, err)

			if !reflect.DeepEqual(tt.want, actual) {
				t.Errorf("want %v, got %v", tt.want, actual)
			}
		})
	}
}
