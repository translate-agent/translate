package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/message/pipeline"
)

// ToGo converts a model.Messages structure into a JSON byte slice
// by first converting it into a format suitable for the pipeline and then encoding it using JSON.
func ToGo(m model.Messages) ([]byte, error) {
	pipelineMsgs := messagesToPipeline(m)

	msg, err := json.Marshal(pipelineMsgs)
	if err != nil {
		return nil, fmt.Errorf("encode pipeline.Messages to JSON: %w", err)
	}

	return msg, nil
}

// FromGo takes a JSON-encoded byte slice, decodes it into a pipeline.Messages structure,
// and then converts it into a model.Messages structure using the messagesFromPipeline function.
func FromGo(b []byte) (model.Messages, error) {
	var pipelineMsgs pipeline.Messages

	if err := json.Unmarshal(b, &pipelineMsgs); err != nil {
		return model.Messages{}, fmt.Errorf("failed to decode JSON to pipeline.Messages: %w", err)
	}

	return messagesFromPipeline(pipelineMsgs), nil
}

// messagesToPipeline converts a model.Messages structure into a pipeline.Messages structure.
func messagesToPipeline(m model.Messages) pipeline.Messages {
	pipelineMsg := pipeline.Messages{
		Language: m.Language,
		Messages: make([]pipeline.Message, 0, len(m.Messages)),
	}

	for _, value := range m.Messages {
		pipelineMsg.Messages = append(pipelineMsg.Messages, pipeline.Message{
			ID:          pipeline.IDList{value.ID},
			Translation: pipeline.Text{Msg: removeEnclosingBrackets(value.Message)},
			Meaning:     value.Description,
			Fuzzy:       value.Fuzzy,
		})
	}

	return pipelineMsg
}

// messagesFromPipeline converts a pipeline.Messages structure into a model.Messages structure.
func messagesFromPipeline(m pipeline.Messages) model.Messages {
	msg := model.Messages{
		Language: m.Language,
		Messages: make([]model.Message, 0, len(m.Messages)),
	}

	for _, value := range m.Messages {
		msg.Messages = append(msg.Messages, model.Message{
			ID:          value.ID[0],
			Fuzzy:       value.Fuzzy,
			Description: value.Meaning,
			Message:     convertToMessageFormatSingular(value.Message.Msg),
		})
	}

	return msg
}
