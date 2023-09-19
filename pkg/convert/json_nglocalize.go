package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// extract-i18n guide and specification: https://angular.io/guide/i18n-common-translation-files
// extract-i18n JSON example: https://github.com/angular/angular/issues/45465

// Extracted messages with json format contains only id:message.
type ngJSON struct {
	Language     language.Tag      `json:"locale"`
	Translations map[string]string `json:"translations"`
}

// FromNgLocalize converts serialized data from the ng extract-i18n tool ("ng extract-i18n --format json")
// into a model.Messages struct.
func FromNgLocalize(data []byte, original bool) (model.Messages, error) {
	var ng ngJSON
	if err := json.Unmarshal(data, &ng); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal @angular/localize JSON into ngJSON struct: %w", err)
	}

	messages := model.Messages{
		Language: ng.Language,
		Messages: make([]model.Message, 0, len(ng.Translations)),
		Original: original,
	}

	status := model.MessageStatusUntranslated
	if original {
		status = model.MessageStatusTranslated
	}

	for k, v := range ng.Translations {
		messages.Messages = append(messages.Messages, model.Message{
			ID:      k,
			Message: convertToMessageFormatSingular(v),
			Status:  status,
		})
	}

	return messages, nil
}

// ToNgLocalize converts a model.Messages struct into a byte slice in @angular/localize JSON format.
func ToNgLocalize(messages model.Messages) ([]byte, error) {
	ng := ngJSON{
		Language:     messages.Language,
		Translations: make(map[string]string, len(messages.Messages)),
	}

	for _, msg := range messages.Messages {
		ng.Translations[msg.ID] = removeEnclosingBrackets(msg.Message)
	}

	data, err := json.Marshal(ng)
	if err != nil {
		return nil, fmt.Errorf("marshal ngJSON struct to @angular/localize JSON : %w", err)
	}

	return data, nil
}
