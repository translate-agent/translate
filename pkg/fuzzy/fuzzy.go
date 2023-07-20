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
	// TODO: only translate messages with untranslated status
	Translate(ctx context.Context, source *model.Messages, targetLanguage language.Tag) (*model.Messages, error)
	// XXX: Method to return supported languages? e.g. SupportedLanguages() map[language.Tag]bool
}

// NoopTranslate implements the Translator interface.
type NoopTranslate struct{}

// Noop Translate returns unmodified incoming messages.
func (n *NoopTranslate) Translate(ctx context.Context, messages *model.Messages) (*model.Messages, error) {
	return messages, nil
}
