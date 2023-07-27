package fuzzy

import (
	"context"
	"fmt"
	"golang.org/x/text/language"
	"strings"

	"go.expect.digital/translate/pkg/model"
)

var SupportedServices = []string{"GoogleTranslate", "AWSTranslate"}

// Usage returns a string describing the supported translators for CLI.
func Usage() string {
	return fmt.Sprintf("translator to use. Supported options: %s", strings.Join(SupportedServices, ", "))
}

type Translator interface {
	Translate(ctx context.Context, messages *model.Messages, targetLangugage language.Tag) (*model.Messages, error)
	// XXX: Method to return supported languages? e.g. SupportedLanguages() map[language.Tag]bool
}

// NoopTranslate implements the Translator interface.
type NoopTranslate struct{}

// Noop Translate returns unmodified incoming messages.
func (n *NoopTranslate) Translate(ctx context.Context, messages *model.Messages, targetLangugage language.Tag) (*model.Messages, error) {
	return messages, nil
}
