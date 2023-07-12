package fuzzy

import (
	"context"

	"go.expect.digital/translate/pkg/model"
)

var SupportedServices = []string{"GoogleTranslate", "AWSTranslate"}

type Translator interface {
	Translate(ctx context.Context, messages *model.Messages) (*model.Messages, error)
	// XXX: Method to return supported languages? e.g. SupportedLanguages() map[language.Tag]bool
}
