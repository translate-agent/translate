package fuzzy

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/translate"
	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// --------------------Definitions--------------------

// Interface that defines some of the methods of the Google Translate client.
// This interface helps to mock the Google Translate client in unit tests.
// https://pkg.go.dev/cloud.google.com/go/translate#Client
type googleClient interface {
	Translate(ctx context.Context, inputs []string, target language.Tag, opts *translate.Options) ([]translate.Translation, error) //nolint:lll
	io.Closer
}

// GoogleTranslate implements the Translator interface.
type GoogleTranslate struct {
	client googleClient
}

type GoogleTranslateOption func(*GoogleTranslate) error

// WithClient sets the Google Translate client.
func WithClient(c googleClient) GoogleTranslateOption {
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

		// Create new Google Cloud service transport with the base of OpenTelemetry HTTP transport.
		trans, err := htransport.NewTransport(
			ctx,
			otelhttp.NewTransport(http.DefaultTransport),
			option.WithAPIKey(apiKey),
		)
		if err != nil {
			return fmt.Errorf("with default client: new transport: %w", err)
		}

		g.client, err = translate.NewClient(ctx, option.WithHTTPClient(&http.Client{Transport: trans}))
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

// --------------------Methods--------------------

func (g *GoogleTranslate) Translate(
	ctx context.Context,
	messages *model.Messages,
	targetLang language.Tag,
) (*model.Messages, error) {
	if messages == nil {
		return nil, nil
	}

	if len(messages.Messages) == 0 {
		return &model.Messages{Language: targetLang}, nil
	}

	// Extract the strings to be send to the Google Translate API.
	targetTexts := make([]string, 0, len(messages.Messages))
	for _, m := range messages.Messages {
		targetTexts = append(targetTexts, m.Message)
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

	translatedMessages := model.Messages{
		Language: targetLang,
		Messages: make([]model.Message, 0, len(translations)),
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
