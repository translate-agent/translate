package translate

import (
	"context"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type TranslationService interface {
	Translate(ctx context.Context, messages *model.Messages, targetLang language.Tag) (*model.Messages, error)
}
