package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var modelMsg = model.Messages{
	Language: language.English,
	Messages: []model.Message{
		{
			ID:          "1",
			Message:     "message1",
			Description: "description1",
			Fuzzy:       true,
		},
		{
			ID:          "2",
			Message:     "message2",
			Description: "description2",
			Fuzzy:       false,
		},
	},
}

func TestToGo(t *testing.T) {
	t.Parallel()

	expected := []byte(`{"language":"en","messages":
[{"id":"1","meaning":"description1","message":"","translation":"message1","fuzzy":true},
{"id":"2","meaning":"description2","message":"","translation":"message2"}]}`)

	buffer := bufferBytes(expected)

	result, err := ToGo(modelMsg)

	assert.NoError(t, err)
	assert.Equal(t, buffer.Bytes(), result)
}

func TestFromGo(t *testing.T) {
	t.Parallel()

	b := []byte(`{"language":"en","messages":
[{"id":"1","meaning":"description1","message":"message1","translation":"","fuzzy":true},
{"id":"2","meaning":"description2","message":"message2","translation":""}]}`)

	buffer := bufferBytes(b)

	result, err := FromGo(buffer.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, modelMsg, result)
}

func bufferBytes(expected []byte) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, expected); err != nil {
		fmt.Println(err)
	}
	return buffer
}
