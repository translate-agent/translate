package fuzzy

import (
	"errors"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var (
	errNilMessages         error = errors.New("nil messages")
	errNoMessages          error = errors.New("no messages")
	errTargetLangUndefined error = errors.New("target language undefined")
)

// validateTranslate validates the given messages and target language
// before executing translation request to save time and resources.
//
// TODO: Add support for validating for source and target language support by the translation service.
func validateTranslate(messages *model.Messages, targetLang language.Tag) error {
	if messages == nil {
		return errNilMessages
	}

	if len(messages.Messages) == 0 {
		return errNoMessages
	}

	if targetLang == language.Und {
		return errTargetLangUndefined
	}

	return nil
}
