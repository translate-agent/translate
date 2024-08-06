package fuzzy

import (
	"context"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var SupportedServices = []string{"GoogleTranslate", "AWSTranslate"}

// Usage returns a string describing the supported translators for CLI.
func Usage() string {
	return "translator to use. Supported options: " + strings.Join(SupportedServices, ", ")
}

type Translator interface {
	Translate(ctx context.Context, translation *model.Translation, targetLanguage language.Tag) (*model.Translation, error)
	// XXX: Method to return supported languages? e.g. SupportedLanguages() map[language.Tag]bool
}

// NoopTranslate implements the Translator interface.
type NoopTranslate struct{}

// Translate returns unmodified incoming translation.
func (n *NoopTranslate) Translate(_ context.Context,
	translation *model.Translation,
	targetLanguage language.Tag,
) (*model.Translation, error) {
	translation.Language = targetLanguage

	return translation, nil
}
