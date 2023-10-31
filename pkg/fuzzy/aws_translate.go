package fuzzy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/translate"
	"github.com/spf13/viper"
	mf "go.expect.digital/translate/pkg/messageformat"
	"go.expect.digital/translate/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/text/language"
)

// --------------------Definitions--------------------

// Interface that defines some of the methods of the AWS Translate client.
// This interface helps to mock the AWS Translate client in unit tests.
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/translate#Client
type awsClient interface {
	TranslateText(
		ctx context.Context,
		params *translate.TranslateTextInput,
		optFns ...func(*translate.Options),
	) (*translate.TranslateTextOutput, error)
}

// AWSTranslate implements the Translator interface.
type AWSTranslate struct {
	client awsClient
}

type AWSTranslateOption func(*AWSTranslate) error

// WithAWSClient sets the AWS Translate client.
func WithAWSClient(awsc awsClient) AWSTranslateOption {
	return func(awst *AWSTranslate) error {
		awst.client = awsc
		return nil
	}
}

// WithDefaultAWSClient creates a new AWS Translate client with credentials from the viper.
func WithDefaultAWSClient(ctx context.Context) AWSTranslateOption {
	return func(awst *AWSTranslate) error {
		accessKey := viper.GetString("other.aws.access_key_id")
		if accessKey == "" {
			return fmt.Errorf("with default client: AWS access key is not set")
		}

		secretKey := viper.GetString("other.aws.secret_access_key")
		if secretKey == "" {
			return fmt.Errorf("with default client: AWS secret key is not set")
		}

		region := viper.GetString("other.aws.region")
		if region == "" {
			return fmt.Errorf("with default client: AWS region is not set")
		}

		// Create a new AWS SDK config
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithHTTPClient(&http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			return fmt.Errorf("load default AWS SDK configuration: %w", err)
		}

		awst.client = translate.NewFromConfig(cfg)

		return nil
	}
}

// NewAWSTranslate creates a new AWS Translate service.
func NewAWSTranslate(ctx context.Context, opts ...AWSTranslateOption) (*AWSTranslate, error) {
	awst := &AWSTranslate{}

	for _, opt := range opts {
		if optErr := opt(awst); optErr != nil {
			return nil, fmt.Errorf("apply opt: %w", optErr)
		}
	}

	// Ping the AWS Translate API to ensure that the client is working.
	_, err := awst.client.TranslateText(ctx, &translate.TranslateTextInput{
		SourceLanguageCode: ptr("en"),
		TargetLanguageCode: ptr("lv"),
		Text:               ptr("Hello World!"),
	})
	if err != nil {
		return nil, fmt.Errorf("AWS translate client: ping AWS translate: %w", err)
	}

	return awst, nil
}

// --------------------Methods--------------------

func (a *AWSTranslate) Translate(ctx context.Context,
	translation *model.Translation,
	targetLanguage language.Tag,
) (*model.Translation, error) {
	if translation == nil {
		return nil, nil
	}

	if len(translation.Messages) == 0 {
		return &model.Translation{Language: targetLanguage, Original: translation.Original}, nil
	}

	// Extract translatable text from translation.
	asts, err := mf.ParseTranslation(translation)
	if err != nil {
		return nil, fmt.Errorf("AWS translate: parse translation messages: %w", err)
	}

	textNodes := mf.GetTextNodes(asts)
	text := textNodes.GetText()
	translatedText := make([]string, 0, len(text))

	for i := range text {
		translateOutput, err := a.client.TranslateText(ctx,
			&translate.TranslateTextInput{
				// Amazon Translate supports text translation between the languages listed in the following table.
				// The language code column uses ISO 639-1 two-digit language codes.
				// For a country variant of a language, the table follows the RFC 5646 format of appending a dash
				// followed by an ISO 3166 2-digit country code.
				// For example, the language code for the Mexican variant of Spanish is es-MX.
				// List of supported languages - https://docs.aws.amazon.com/translate/latest/dg/what-is-languages.html

				TargetLanguageCode: awsLanguage(targetLanguage),
				SourceLanguageCode: awsLanguage(translation.Language),
				// Maximum text size limit accepted by the AWS Translate API - 10000 bytes.
				Text: ptr(text[i]),
			})
		if err != nil {
			return nil, fmt.Errorf("AWS translate: translate text #%d: %w", i, err)
		}

		translatedText = append(translatedText, *translateOutput.TranslatedText)
	}

	// Overwrite text nodes in ASTs to include newly translated text.
	if err := textNodes.OverwriteText(translatedText); err != nil {
		return nil, fmt.Errorf("AWS translate: overwrite text nodes in ASTs: %w", err)
	}

	// create translation with newly translated messages.
	translated := model.Translation{
		Language: targetLanguage,
		Original: translation.Original,
		Messages: make([]model.Message, len(translation.Messages)),
	}

	for i := range asts {
		var err error

		m := translation.Messages[i]

		if m.Message, err = mf.Compile(asts[i]); err != nil {
			return nil, fmt.Errorf("AWS translate: compile AST '#%d': %w", i, err)
		}

		m.Status = model.MessageStatusFuzzy
		translated.Messages[i] = m
	}

	return &translated, nil
}

// helpers

// ptr returns pointer to the passed in value.
func ptr[T any](v T) *T {
	return &v
}

// awsLanguage normalizes language.Tag to be usable by AWS translate.
// skips locale part if region is not a country,
// AWS only supports ISO 3166 2-digit country codes.
// https://docs.aws.amazon.com/translate/latest/dg/what-is-languages.html
func awsLanguage(language language.Tag) *string {
	lang := language.String()

	if region, _ := language.Region(); !region.IsCountry() {
		baseLanguage, _ := language.Base()
		lang = baseLanguage.String()
	}

	return &lang
}
