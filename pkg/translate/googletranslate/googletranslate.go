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
// This interface helps to mock the Google Translate client in unit tests.
// https://pkg.go.dev/cloud.google.com/go/translate#Client
type GoogleClient interface {
	Translate(ctx context.Context, inputs []string, target language.Tag, opts *translate.Options) ([]translate.Translation, error) //nolint:lll
	io.Closer
}

// GoogleTranslate implements the TranslationService interface.
type GoogleTranslate struct {
	client GoogleClient
}

type GoogleTranslateOption func(*GoogleTranslate) error

// WithClient sets the Google Translate client.
func WithClient(c GoogleClient) GoogleTranslateOption {
	return func(g *GoogleTranslate) error {
		g.client = c
		return nil
	}
}

// WithDefaultClient creates a new Google Translate client with the API key from the viper.
func WithDefaultClient(ctx context.Context) GoogleTranslateOption {
	return func(g *GoogleTranslate) error {
		var err error

		apiKey := viper.GetString("translate_services.google_translate.api_key")
		if apiKey == "" {
			return fmt.Errorf("with default client: google translate api key is not set")
		}

		g.client, err = translate.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			return fmt.Errorf("with default client: new google translate client: %w", err)
		}

		return nil
	}
}

// NewGoogleTranslate creates a new Google Translate service.
func NewGoogleTranslate(
	ctx context.Context,
	opts ...GoogleTranslateOption,
) (gt *GoogleTranslate, closer func() error, err error) {
	gt = &GoogleTranslate{}

	for _, opt := range opts {
		if optErr := opt(gt); optErr != nil {
			return nil, nil, fmt.Errorf("apply opt: %w", optErr)
		}
	}

	// Ping the Google Translate API to ensure that the client is working.
	_, err = gt.client.Translate(ctx, []string{"Hello World!"}, language.Latvian, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("google translate client: ping google translate: %w", err)
	}

	return gt, gt.client.Close, nil
}
