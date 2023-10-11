package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/message/pipeline"
)

// ToGo converts a model.Translation structure into a JSON byte slice
// by first converting it into a format suitable for the pipeline and then encoding it using JSON.
func ToGo(t model.Translation) ([]byte, error) {
	pipelineMsgs := translationToPipeline(t)

	msg, err := json.Marshal(pipelineMsgs)
	if err != nil {
		return nil, fmt.Errorf("encode pipeline.Messages to JSON: %w", err)
	}

	return msg, nil
}

// FromGo takes a JSON-encoded byte slice, decodes it into a pipeline.Messages structure,
// and then converts it into a model.Translation structure using the translationFromPipeline function.
func FromGo(b []byte, original bool) (model.Translation, error) {
	var pipelineMsgs pipeline.Messages

	if err := json.Unmarshal(b, &pipelineMsgs); err != nil {
		return model.Translation{}, fmt.Errorf("decode JSON to pipeline.Messages: %w", err)
	}

	return translationFromPipeline(pipelineMsgs, original), nil
}

// translationToPipeline converts a model.Translation structure into a pipeline.Messages structure.
func translationToPipeline(t model.Translation) pipeline.Messages {
	pipelineMsg := pipeline.Messages{
		Language: t.Language,
		Messages: make([]pipeline.Message, 0, len(t.Messages)),
	}

	for _, value := range t.Messages {
		msg := pipeline.Message{
			ID:          pipeline.IDList{value.ID},
			Translation: pipeline.Text{Msg: removeEnclosingBrackets(value.Message)},
			Meaning:     value.Description,
			Fuzzy:       value.Status == model.MessageStatusFuzzy,
		}

		switch len(value.Positions) {
		default:
			for _, pos := range value.Positions {
				msg.Position = pos
				pipelineMsg.Messages = append(pipelineMsg.Messages, msg)
			}
		case 0:
			pipelineMsg.Messages = append(pipelineMsg.Messages, msg)
		}
	}

	return pipelineMsg
}

// translationFromPipeline converts a pipeline.Messages structure into a model.Translation structure.
func translationFromPipeline(m pipeline.Messages, original bool) model.Translation {
	translation := model.Translation{
		Language: m.Language,
		Messages: make([]model.Message, 0, len(m.Messages)),
		Original: original,
	}

	getMessage := func(m pipeline.Message) string { return m.Translation.Msg }
	getStatus := func(m pipeline.Message) model.MessageStatus {
		if m.Fuzzy {
			return model.MessageStatusFuzzy
		}

		return model.MessageStatusUntranslated
	}

	if original {
		getMessage = func(m pipeline.Message) string { return m.Message.Msg }
		getStatus = func(_ pipeline.Message) model.MessageStatus { return model.MessageStatusTranslated }
	}

	for _, value := range m.Messages {
		msg := model.Message{
			ID:          value.ID[0],
			Description: value.Meaning,
			Message:     convertToMessageFormatSingular(getMessage(value)),
			Status:      getStatus(value),
		}

		if value.Position != "" {
			msg.Positions = []string{value.Position}
		}

		translation.Messages = append(translation.Messages, msg)
	}

	return translation
}
