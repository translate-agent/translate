package googletranslate

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/translate"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func (g *GoogleTranslate) Translate(
	ctx context.Context,
	messages *model.Messages,
	targetLang language.Tag,
) (*model.Messages, error) {
	if messages == nil {
		return nil, errors.New("google translate: translate: nil messages")
	}

	if len(messages.Messages) == 0 {
		return messages, errors.New("google translate: translate: no messages")
	}

	if targetLang == language.Und {
		return nil, errors.New("google translate: translate: target language undefined")
	}

	textsToTranslate := make([]string, 0, len(messages.Messages))
	for _, m := range messages.Messages {
		textsToTranslate = append(textsToTranslate, m.ID)
	}

	// Set source language if defined, otherwise let Google Translate detect it.
	opts := &translate.Options{}
	if messages.Language != language.Und {
		opts = &translate.Options{Source: messages.Language}
	}

	translations, err := g.client.Translate(ctx, textsToTranslate, targetLang, opts)
	if err != nil {
		return nil, fmt.Errorf("google translate client: translate: %w", err)
	}

	translatedMessages := make([]model.Message, 0, len(translations))
	for i, t := range translations {
		translatedMessages = append(translatedMessages, model.Message{
			ID:          messages.Messages[i].ID,
			PluralID:    messages.Messages[i].PluralID,
			Description: messages.Messages[i].Description,
			Message:     t.Text,
			Fuzzy:       true,
		})
	}

	return &model.Messages{Language: messages.Language, Messages: translatedMessages}, nil
}
