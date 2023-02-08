package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/message/pipeline"
)

func ToGo(m model.Messages) ([]byte, error) {
	pipelineMsg := messagesToPipeline(m)

	msg, err := json.Marshal(pipelineMsg)
	if err != nil {
		return nil, fmt.Errorf("encode pipeline.Messages to JSON: %w", err)
	}

	return msg, nil
}

func FromGo(b []byte) (model.Messages, error) {
	var pipelineMsg pipeline.Messages

	err := json.Unmarshal(b, &pipelineMsg)
	if err != nil {
		return model.Messages{}, fmt.Errorf("decode JSON to pipeline.Messages: %w", err)
	}

	msg := messagesFromPipeline(pipelineMsg)

	return msg, nil
}

func messagesToPipeline(m model.Messages) pipeline.Messages {
	pipelineMsg := pipeline.Messages{
		Language: m.Language,
		Messages: make([]pipeline.Message, 0, len(m.Messages)),
	}

	for _, value := range m.Messages {
		pipelineMsg.Messages = append(pipelineMsg.Messages, pipeline.Message{
			ID:      pipeline.IDList{value.ID},
			Message: pipeline.Text{Msg: value.Message},
			Meaning: value.Description,
			Fuzzy:   value.Fuzzy,
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
			Message:     value.Translation.Msg,
		})
	}

	return msg
}
