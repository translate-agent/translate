package fuzzy

import (
	"context"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var SupportedServices = []string{"GoogleTranslate"}

type TranslationService interface {
	Translate(ctx context.Context, messages *model.Messages, targetLang language.Tag) (*model.Messages, error)
	// XXX: Method to return supported languages? e.g. SupportedLanguages() map[language.Tag]bool
}
