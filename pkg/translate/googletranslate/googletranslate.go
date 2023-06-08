package googletranslate

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/translate"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// Interface that defines some of the methods of the Google Translate client.
// This interface helps to mock the Google Translate client in tests.
// https://pkg.go.dev/cloud.google.com/go/translate#Client
type GoogleClient interface {
	Translate(ctx context.Context, inputs []string, target language.Tag, opts *translate.Options) ([]translate.Translation, error) //nolint:lll
	SupportedLanguages(ctx context.Context, target language.Tag) ([]translate.Language, error)
	io.Closer
}

// GoogleTranslate implements the TranslationService interface.
type GoogleTranslate struct {
	client            GoogleClient
	supportedLangTags map[language.Tag]bool
}

type TranslateOption func(*GoogleTranslate) error

// WithClient sets the Google Translate client.
func WithClient(c GoogleClient) TranslateOption {
	return func(g *GoogleTranslate) error {
		g.client = c
		return nil
	}
}

// WithDefaultClient creates a new Google Translate client with the API key from the viper.
func WithDefaultClient(ctx context.Context) TranslateOption {
	return func(g *GoogleTranslate) error {
		var err error

		g.client, err = translate.NewClient(ctx, option.WithAPIKey(viper.GetString("googletranslate.api.key")))
		if err != nil {
			return fmt.Errorf("with default client: new google translate client: %w", err)
		}

		return nil
	}
}

func NewGoogleTranslate(ctx context.Context, opts ...TranslateOption) (*GoogleTranslate, func() error, error) {
	googleTranslate := &GoogleTranslate{}

	for _, opt := range opts {
		if err := opt(googleTranslate); err != nil {
			return nil, nil, fmt.Errorf("apply opt: %w", err)
		}
	}

	// Ping the Google Translate API to ensure that the client is working.
	_, err := googleTranslate.client.Translate(ctx, []string{"Hello World!"}, language.Latvian, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("google translate client: ping google translate: %w", err)
	}

	// Get the list of supported languages.
	supported, err := googleTranslate.client.SupportedLanguages(ctx, language.English)
	if err != nil {
		return nil, nil, fmt.Errorf("google translate client: get supported languages: %w", err)
	}

	// Create a map of supported languages for quick lookup.
	googleTranslate.supportedLangTags = make(map[language.Tag]bool, len(supported))
	for _, lang := range supported {
		googleTranslate.supportedLangTags[lang.Tag] = true
	}

	// Add the undefined language tag to the map of supported languages
	// as Google Translate tries to detect the language then.
	googleTranslate.supportedLangTags[language.Und] = true

	return googleTranslate, googleTranslate.client.Close, nil
}
