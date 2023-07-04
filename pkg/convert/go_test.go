package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func TestToGo(t *testing.T) {
	t.Parallel()

	modelMsg := model.Messages{
		Language: language.English,
		Messages: []model.Message{
			{
				ID:          "1",
				Message:     "{message1}",
				Description: "description1",
				Positions:   []string{"src/config.go:10", "src/config.go:20"},
				Fuzzy:       true,
			},
			{
				ID:          "2",
				Message:     "{message2}",
				Description: "description2",
				Fuzzy:       false,
			},
		},
	}

	expected := []byte(`
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
	}`)

	actual, err := ToGo(modelMsg)
	if !assert.NoError(t, err) {
		return
	}

	assert.JSONEq(t, string(expected), string(actual))
}

func TestFromGo(t *testing.T) {
	t.Parallel()

	modelMsg := model.Messages{
		Language: language.English,
		Messages: []model.Message{
			{
				ID:          "1",
				Message:     "{message1}",
				Description: "description1",
				Positions:   []string{"src/config.go:10"},
				Fuzzy:       true,
			},
			{
				ID:          "1",
				Message:     "{message1}",
				Description: "description1",
				Positions:   []string{"src/config.go:20"},
				Fuzzy:       true,
			},
			{
				ID:          "2",
				Message:     "{message2}",
				Description: "description2",
				Fuzzy:       false,
			},
		},
	}

	input := []byte(`
	{
		"language":"en",
		"messages":[
			{
				"id":"1",
				"meaning":"description1",
				"message":"message1",
				"translation":"",
				"position": "src/config.go:10",
				"fuzzy":true
			},
			{
				"id":"1",
				"meaning":"description1",
				"message":"message1",
				"translation":"",
				"position": "src/config.go:20",
				"fuzzy":true
			},
			{
				"id":"2",
				"meaning":"description2",
				"message":"message2",
				"translation":""
			}
		]
	}`)

	actual, err := FromGo(input)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, modelMsg, actual)
}
