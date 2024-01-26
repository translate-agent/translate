package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/mf2"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// extract-i18n guide and specification: https://angular.io/guide/i18n-common-translation-files
// extract-i18n JSON example: https://github.com/angular/angular/issues/45465

// Extracted translation with json format contains only id:message.
type ngJSON struct {
	Language     language.Tag      `json:"locale"`
	Translations map[string]string `json:"translations"`
}

// FromNgLocalize converts serialized data from the ng extract-i18n tool ("ng extract-i18n --format json")
// into a model.Translation struct.
func FromNgLocalize(data []byte, original *bool) (model.Translation, error) {
	var ng ngJSON
	if err := json.Unmarshal(data, &ng); err != nil {
		return model.Translation{}, fmt.Errorf("unmarshal @angular/localize JSON into ngJSON struct: %w", err)
	}

	// if original is not provided default to false.
	if original == nil {
		original = ptr(false)
	}

	translation := model.Translation{
		Language: ng.Language,
		Messages: make([]model.Message, 0, len(ng.Translations)),
		Original: *original,
	}

	status := model.MessageStatusUntranslated
	if *original {
		status = model.MessageStatusTranslated
	}

	for k, v := range ng.Translations {
		val, err := toMF2(v)
		if err != nil {
			return model.Translation{}, fmt.Errorf("convert string to MF2: %w", err)
		}

		translation.Messages = append(translation.Messages, model.Message{
			ID:      k,
			Message: val,
			Status:  status,
		})
	}

	return translation, nil
}

func toMF2(msg string) (string, error) {
	b := mf2.NewBuilder()

	b.Text(msg)

	s, err := b.Build()
	if err != nil {
		return "", fmt.Errorf("build MF2: %w", err)
	}

	return s, nil
}

// ToNgLocalize converts a model.Translation struct into a byte slice in @angular/localize JSON format.
func ToNgLocalize(translation model.Translation) ([]byte, error) {
	ng := ngJSON{
		Language:     translation.Language,
		Translations: make(map[string]string, len(translation.Messages)),
	}

	for _, msg := range translation.Messages {
		ng.Translations[msg.ID] = "" // TODO: convert msg.Message from MF2 format.
	}

	data, err := json.Marshal(ng)
	if err != nil {
		return nil, fmt.Errorf("marshal ngJSON struct to @angular/localize JSON : %w", err)
	}

	return data, nil
}
