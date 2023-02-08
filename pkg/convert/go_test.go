package convert

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
	"golang.org/x/text/message/pipeline"
)

var inputMsg = model.Messages{
	Language: language.English,
	Messages: []model.Message{
		{
			ID:          "1",
			Message:     "hello",
			Description: "greeting",
			Fuzzy:       false,
		},
		{
			ID:          "2",
			Message:     "bye",
			Description: "greeting",
			Fuzzy:       false,
		},
	},
}

var pipelineMsg = pipeline.Messages{
	Language: language.English,
	Messages: []pipeline.Message{
		{
			ID:      []string{"1"},
			Message: pipeline.Text{Msg: "hello"},
			Meaning: "greeting",
			Fuzzy:   false,
		},
		{
			ID:      []string{"2"},
			Message: pipeline.Text{Msg: "welcome"},
			Meaning: "greeting",
			Fuzzy:   false,
		},
	},
}

var (
	expectedMessages      = messagesFromPipeline(pipelineMsg)
	expectedByteResult, _ = json.Marshal(messagesToPipeline(inputMsg)) //nolint:errchkjson
)

func Test_ToGo_ReturnsExpectedByteSlice(t *testing.T) {
	t.Parallel()

	actualByteResult, err := ToGo(inputMsg)
	if err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, expectedByteResult, actualByteResult)
}

func Test_FromGo_ReturnsExpectedModel(t *testing.T) {
	t.Parallel()

	actualMessages, err := FromGo(expectedByteResult)
	if err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, expectedMessages, actualMessages)
}
