package fuzzy

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	awstranslate "github.com/aws/aws-sdk-go-v2/service/translate"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"

	"cloud.google.com/go/translate"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// --------------------Definitions--------------------

type awsClient interface {
	Translate(context.Context, []string, language.Tag, *awstranslate.Options) ([]translate.Translation, error)
	Ping(context.Context) error
	io.Closer
}

// AWSTranslate implements the Translator interface.
type AWSTranslate struct {
	client awsClient
}

type AWSTranslateOption func(*AWSTranslate) error

// WithAWSClient sets the AWS Translate client.
func WithAWSClient(c awsClient) AWSTranslateOption {
	return func(a *AWSTranslate) error {
		a.client = c
		return nil
	}
}

// WithAWSDefaultClient creates a new AWS Translate client with credentials from the viper.
func WithDefaultAWSClient(ctx context.Context) AWSTranslateOption {
	return func(awst *AWSTranslate) error {
		// Create a new AWS SDK config
		cfg, err := config.LoadDefaultConfig(ctx, config.WithHTTPClient())
		if err != nil {
			return fmt.Errorf("failed to load AWS SDK configuration:", err)
		}

		// Create new AWS translate service transport with the base of OpenTelemetry HTTP transport.
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

// NewAWSTranslate creates a new AWS Translate service.
func NewAWSTranslate(ctx context.Context, opts ...AWSTranslateOption) (
	awst *AWSTranslate, closer func() error, err error) {
	awst = &AWSTranslate{}

	for _, opt := range opts {
		if optErr := opt(awst); optErr != nil {
			return nil, nil, fmt.Errorf("apply opt: %w", optErr)
		}
	}

	// Ping the AWS Translate API to ensure that the client is working.
	if err := awst.client.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("aws translate client: ping aws translate: %w", err)
	}

	return awst, awst.client.Close, nil
}

// --------------------Methods--------------------

func (a *AWSTranslate) Translate(ctx context.Context, messages *model.Messages) (*model.Messages, error) {
	if messages == nil {
		return nil, nil
	}

	if len(messages.Messages) == 0 {
		return &model.Messages{Language: messages.Language, Original: messages.Original}, nil
	}

	// Extract the strings to be send to the AWS Translate API.
	targetTexts := make([]string, 0, len(messages.Messages))
	for _, m := range messages.Messages {
		targetTexts = append(targetTexts, m.Message)
	}

	// Create a new AWS SDK config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Failed to load AWS SDK configuration:", err)
		return
	}

	client := awstranslate.NewFromConfig(cfg)

	client.StartTextTranslationJob()

	client.StartTextTranslationJob()

	translations, err := g.client.Translate(ctx, targetTexts, messages.Language, nil)
	if err != nil {
		return nil, fmt.Errorf("google translate client: translate: %w", err)
	}

	translatedMessages := model.Messages{
		Language: messages.Language,
		Original: messages.Original,
		Messages: make([]model.Message, 0, len(translations)),
	}

	for i, t := range translations {
		translatedMessages.Messages = append(translatedMessages.Messages, model.Message{
			ID:          messages.Messages[i].ID,
			PluralID:    messages.Messages[i].PluralID,
			Description: messages.Messages[i].Description,
			Message:     t.Text,
			Status:      model.MessageStatusFuzzy,
			Positions:   messages.Messages[i].Positions,
		})
	}

	return &translatedMessages, nil
}
