package googletranslate

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/translate"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func (g *GoogleTranslate) validateTranslateReq(messages *model.Messages, targetLang language.Tag) error {
	if messages == nil {
		return errors.New("nil messages")
	}

	if len(messages.Messages) == 0 {
		return errors.New("no messages")
	}

	if targetLang == language.Und {
		return errors.New("target language undefined")
	}

	// Enforce that source language is supported.
	if ok := g.supportedLangTags[messages.Language]; !ok {
		return fmt.Errorf("source language %s not supported", messages.Language)
	}

	// Enforce that target language is supported.
	if ok := g.supportedLangTags[targetLang]; !ok {
		return fmt.Errorf("target language %s not supported", targetLang)
	}

	return nil
}

func (g *GoogleTranslate) Translate(
	ctx context.Context,
	messages *model.Messages,
	targetLang language.Tag,
) (*model.Messages, error) {
	if err := g.validateTranslateReq(messages, targetLang); err != nil {
		return nil, fmt.Errorf("google translate: validate translate request: %w", err)
	}

	targetTexts := make([]string, 0, len(messages.Messages))
	for _, m := range messages.Messages {
		targetTexts = append(targetTexts, m.ID)
	}

	// Set source language if defined, otherwise let Google Translate detect it.
	opts := &translate.Options{}
	if messages.Language != language.Und {
		opts = &translate.Options{Source: messages.Language}
	}

	translations, err := g.client.Translate(ctx, targetTexts, targetLang, opts)
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
