package convert

import (
	"encoding/json"
	"fmt"

	ast "go.expect.digital/mf2/parse"

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
		msg, err := mf2.NewBuilder().Text(v).Build()
		if err != nil {
			return model.Translation{}, fmt.Errorf("convert string to MF2: %w", err)
		}

		translation.Messages = append(translation.Messages, model.Message{
			ID:      k,
			Message: msg,
			Status:  status,
		})
	}

	return translation, nil
}

// ToNgLocalize converts a model.Translation struct into a byte slice in @angular/localize JSON format.
func ToNgLocalize(translation model.Translation) ([]byte, error) {
	ng := ngJSON{
		Language:     translation.Language,
		Translations: make(map[string]string, len(translation.Messages)),
	}

	for _, msg := range translation.Messages {
		tree, err := ast.Parse(msg.Message)
		if err != nil {
			return nil, fmt.Errorf("parse mf2 message: %w", err)
		}

		switch mf2Msg := tree.Message.(type) {
		case nil:
			ng.Translations[msg.ID] = ""
		case ast.SimpleMessage:
			ng.Translations[msg.ID] = patternsToSimpleMsg(mf2Msg)
		case ast.ComplexMessage:
			return nil, fmt.Errorf("complex message not supported")
		}
	}

	data, err := json.Marshal(ng)
	if err != nil {
		return nil, fmt.Errorf("marshal ngJSON struct to @angular/localize JSON : %w", err)
	}

	return data, nil
}
