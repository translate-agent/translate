package googletranslate

import (
	"context"
	"fmt"

	gtranslate "cloud.google.com/go/translate"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/translate/common"
	"golang.org/x/text/language"
)

func (g *GoogleTranslate) Translate(
	ctx context.Context,
	messages *model.Messages,
	targetLang language.Tag,
) (*model.Messages, error) {
	if err := common.ValidateTranslate(messages, targetLang); err != nil {
		return nil, fmt.Errorf("google translate: validate translate request: %w", err)
	}

	translatedMessages := model.Messages{
		Language: targetLang,
		Messages: make([]model.Message, 0, len(messages.Messages)),
	}

	targetTexts := make([]string, 0, len(messages.Messages))
	for _, m := range messages.Messages {
		targetTexts = append(targetTexts, m.Message)
	}

	// Set source language if defined, otherwise let Google Translate detect it.
	opts := &gtranslate.Options{}
	if messages.Language != language.Und {
		opts = &gtranslate.Options{Source: messages.Language}
	}

	translations, err := g.client.Translate(ctx, targetTexts, targetLang, opts)
	if err != nil {
		return nil, fmt.Errorf("google translate client: translate: %w", err)
	}

	for i, t := range translations {
		translatedMessages.Messages = append(translatedMessages.Messages, model.Message{
			ID:          messages.Messages[i].ID,
			PluralID:    messages.Messages[i].PluralID,
			Description: messages.Messages[i].Description,
			Message:     t.Text,
			Fuzzy:       true,
		})
	}

	return &translatedMessages, nil
}
