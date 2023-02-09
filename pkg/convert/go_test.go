package convert

import (
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

	test := struct {
		m    model.Messages
		want []byte
	}{
		m:    modelMsg,
		want: []byte(`{"language":"en","messages":[{"id":"1","meaning":"description1","message":"","translation":"message1","fuzzy":true},{"id":"2","meaning":"description2","message":"","translation":"message2"}]}`), //nolint:lll
	}

	result, err := ToGo(test.m)

	assert.NoError(t, err)
	assert.Equal(t, test.want, result)
}

func TestFromGo(t *testing.T) {
	t.Parallel()

	test := struct {
		m    []byte
		want model.Messages
	}{
		m: []byte(`{"language":"en","messages":[{"id":"1","meaning":"description1","message":"message1","translation":"","fuzzy":true},{"id":"2","meaning":"description2","message":"message2","translation":""}]}`), //nolint:lll

		want: modelMsg,
	}
	result, err := FromGo(test.m)

	assert.NoError(t, err)
	assert.Equal(t, test.want, result)
}
