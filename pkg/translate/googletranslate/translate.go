package googletranslate

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

type GoogleClient interface {
	Translate(ctx context.Context, inputs []string, target language.Tag, opts *translate.Options) ([]translate.Translation, error) //nolint:lll
	SupportedLanguages(ctx context.Context, target language.Tag) ([]translate.Language, error)
	io.Closer
}

type GoogleTranslate struct {
	client            GoogleClient
	supportedLangTags map[language.Tag]bool
}

func NewGoogleTranslate(ctx context.Context, c GoogleClient) (*GoogleTranslate, error) {
	// Ping the Google Translate API to ensure that the client is working.
	_, err := c.Translate(ctx, []string{"Hello World!"}, language.Latvian, nil)
	if err != nil {
		return nil, fmt.Errorf("new google translate client: ping google translate: %w", err)
	}

	// Get the list of supported languages.
	supported, err := c.SupportedLanguages(ctx, language.English)
	if err != nil {
		return nil, fmt.Errorf("new google translate client: get supported languages: %w", err)
	}

	// Create a map of supported languages for quick lookup.
	supportedLangTags := make(map[language.Tag]bool, len(supported))
	for _, lang := range supported {
		supportedLangTags[lang.Tag] = true
	}

	// Add the undefined language tag to the map of supported languages
	// as Google Translate tries to detect the language then.
	supportedLangTags[language.Und] = true

	return &GoogleTranslate{client: c, supportedLangTags: supportedLangTags}, nil
}
