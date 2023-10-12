package fuzzy

import (
	"context"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var SupportedServices = []string{"GoogleTranslate", "AWSTranslate"}

// Usage returns a string describing the supported translators for CLI.
func Usage() string {
	return fmt.Sprintf("translator to use. Supported options: %s", strings.Join(SupportedServices, ", "))
}

type Translator interface {
	Translate(ctx context.Context, translation *model.Translation, targetLanguage language.Tag) (*model.Translation, error)
	// XXX: Method to return supported languages? e.g. SupportedLanguages() map[language.Tag]bool
}

// NoopTranslate implements the Translator interface.
type NoopTranslate struct{}

// Translate returns unmodified incoming translation.
func (n *NoopTranslate) Translate(ctx context.Context,
	translation *model.Translation,
	targetLanguage language.Tag,
) (*model.Translation, error) {
	translation.Language = targetLanguage

	return translation, nil
}
