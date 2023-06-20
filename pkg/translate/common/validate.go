package common

import (
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// ValidateTranslate validates the given messages and target language
// before executing translation request to save time and resources.
//
// TODO: Add support for validating for source and target language support by the translation service.
func ValidateTranslate(messages *model.Messages, targetLang language.Tag) error {
	if messages == nil {
		return ErrNilMessages
	}

	if len(messages.Messages) == 0 {
		return ErrNoMessages
	}

	if targetLang == language.Und {
		return ErrTargetLangUndefined
	}

	return nil
}
