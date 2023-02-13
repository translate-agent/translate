package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/message/pipeline"
)

func ToGo(m model.Messages) ([]byte, error) {
	pipelineMsgs := messagesToPipeline(m)

	msg, err := json.Marshal(pipelineMsgs)
	if err != nil {
		return nil, fmt.Errorf("encode pipeline.Messages to JSON: %w", err)
	}

	return msg, nil
}

func FromGo(b []byte) (model.Messages, error) {
	var pipelineMsgs pipeline.Messages

	if err := json.Unmarshal(b, &pipelineMsgs); err != nil {
		return model.Messages{}, fmt.Errorf("failed to decode JSON to pipeline.Messages: %w", err)
	}

	return messagesFromPipeline(pipelineMsgs), nil
}

func messagesToPipeline(m model.Messages) pipeline.Messages {
	pipelineMsg := pipeline.Messages{
		Language: m.Language,
		Messages: make([]pipeline.Message, 0, len(m.Messages)),
	}

	for _, value := range m.Messages {
		pipelineMsg.Messages = append(pipelineMsg.Messages, pipeline.Message{
			ID:          pipeline.IDList{value.ID},
			Translation: pipeline.Text{Msg: value.Message},
			Meaning:     value.Description,
			Fuzzy:       value.Fuzzy,
		})
	}

	return pipelineMsg
}

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
			Message:     value.Message.Msg,
		})
	}

	return msg
}
