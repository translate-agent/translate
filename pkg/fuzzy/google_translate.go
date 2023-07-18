package fuzzy

import (
	"context"
	"fmt"
	"github.com/googleapis/gax-go/v2"
	"net/http"
	"strings"

	translatev3 "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// --------------------Definitions--------------------

// Interface that defines some of the methods of the Google Translate client.
type googleClient interface {
	TranslateText(ctx context.Context, req *translatepb.TranslateTextRequest, opts ...gax.CallOption) (*translatepb.TranslateTextResponse, error)
	Close() error
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

		apiKey := viper.GetString("other.google_translate.api_key")
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

		g.client, err = translatev3.NewTranslationClient(ctx, option.WithHTTPClient(&http.Client{Transport: trans}))
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
	_, err = gt.client.TranslateText(ctx, &translatepb.TranslateTextRequest{
		SourceLanguageCode: "en",
		TargetLanguageCode: "lv",
		Contents:           []string{"Hello World!"},
		Parent:             fmt.Sprintf("projects/%s/locations/%s", viper.GetString("other.google_translate.project_id"), "global"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("google translate client: ping google translate: %w", err)
	}

	return gt, gt.client.Close, nil
}

// --------------------Methods--------------------

func (g *GoogleTranslate) Translate(
	ctx context.Context,
	messages *model.Messages,
) (*model.Messages, error) {
	if messages == nil {
		return nil, nil
	}

	if len(messages.Messages) == 0 {
		return &model.Messages{Language: messages.Language, Original: messages.Original}, nil
	}

	// Extract the strings to be send to the Google Translate API.
	targetTexts := make([]string, 0, len(messages.Messages))
	for _, m := range messages.Messages {
		targetTexts = append(targetTexts, m.Message)
	}

	resp, err := g.client.TranslateText(ctx, &translatepb.TranslateTextRequest{
		SourceLanguageCode: "en", // specify your source language if it's not English
		TargetLanguageCode: strings.ToLower(messages.Language.String()),
		Contents:           targetTexts,
		Parent:             fmt.Sprintf("projects/%s/locations/%s", viper.GetString("other.google_translate.project_id"), "global"),
	})
	if err != nil {
		return nil, fmt.Errorf("google translate client: translate: %w", err)
	}

	translatedMessages := model.Messages{
		Language: messages.Language,
		Original: messages.Original,
		Messages: make([]model.Message, 0, len(resp.GetTranslations())),
	}

	for i, t := range resp.GetTranslations() {
		translatedMessages.Messages = append(translatedMessages.Messages, model.Message{
			ID:          messages.Messages[i].ID,
			PluralID:    messages.Messages[i].PluralID,
			Description: messages.Messages[i].Description,
			Message:     t.GetTranslatedText(),
			Status:      model.MessageStatusFuzzy,
			Positions:   messages.Messages[i].Positions,
		})
	}

	return &translatedMessages, nil
}
