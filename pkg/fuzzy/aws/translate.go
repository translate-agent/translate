package awstranslate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/translate"

	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/model"
)

// --------------------Definitions--------------------

type awsClient interface {
	TranslateText(
		ctx context.Context,
		params *translate.TranslateTextInput,
		optFns ...func(*translate.Options),
	) (*translate.TranslateTextOutput, error)
}

// Translate implements the Translator interface.
type Translate struct {
	client awsClient
}

type TranslateOption func(*Translate) error

// WithClient sets the AWS Translate client.
func WithClient(ac awsClient) TranslateOption {
	return func(tr *Translate) error {
		tr.client = ac
		return nil
	}
}

// WithDefaultClient creates a new AWS Translate client with credentials from the viper.
func WithDefaultClient(ctx context.Context) TranslateOption {
	return func(tr *Translate) error {
		accessKey := viper.GetString("other.aws_translate.access_key")
		if accessKey == "" {
			return fmt.Errorf("with default client: aws translate access key is not set")
		}

		secretKey := viper.GetString("other.aws_translate.secret_key")
		if secretKey == "" {
			return fmt.Errorf("with default client: aws translate secret key is not set")
		}

		region := viper.GetString("other.aws_translate.region")
		if region == "" {
			return fmt.Errorf("with default client: aws translate region is not set")
		}

		// Create a new AWS SDK config
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithHTTPClient(http.DefaultClient),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			return fmt.Errorf("failed to load default AWS SDK configuration: %w", err)
		}

		tr.client = translate.NewFromConfig(cfg)

		return nil
	}
}

// NewTranslate creates a new AWS Translate service.
func NewTranslate(ctx context.Context, opts ...TranslateOption) (*Translate, error) {
	tr := &Translate{}

	for _, opt := range opts {
		if optErr := opt(tr); optErr != nil {
			return nil, fmt.Errorf("apply opt: %w", optErr)
		}
	}

	// Ping the AWS Translate API to ensure that the client is working.
	_, err := tr.client.TranslateText(ctx, &translate.TranslateTextInput{
		SourceLanguageCode: ptr("en"),
		TargetLanguageCode: ptr("lv"),
		Text:               ptr("Hello World!"),
	})
	if err != nil {
		return nil, fmt.Errorf("aws translate client: ping aws translate: %w", err)
	}

	return tr, nil
}

// Maximum text size limit accepted by the AWS Translate API - 10000 bytes.
const inputTextSizeLimit = 10_000

// --------------------Methods--------------------

func (tr *Translate) Translate(ctx context.Context, messages *model.Messages) (*model.Messages, error) {
	if messages == nil {
		return nil, nil
	}

	if len(messages.Messages) == 0 {
		return &model.Messages{Language: messages.Language, Original: messages.Original}, nil
	}

	var (
		buf  bytes.Buffer
		bufs []bytes.Buffer
	)

	for _, m := range messages.Messages {
		if buf.Len()+len(m.Message) > inputTextSizeLimit {
			bufs = append(bufs, buf)
			buf = bytes.Buffer{}
		}

		fmt.Fprintln(&buf, m.Message)
	}

	bufs = append(bufs, buf)

	translatedTexts := make([]string, 0, len(messages.Messages))

	// skip locale part if region is not a country,
	// AWS only supports ISO 3166 2-digit country codes.
	targetLanguage := messages.Language.String()

	if region, _ := messages.Language.Region(); !region.IsCountry() {
		baseLang, _ := messages.Language.Base()
		targetLanguage = baseLang.String()
	}

	for i := range bufs {
		translateOutput, err := tr.client.TranslateText(ctx,
			&translate.TranslateTextInput{
				// Amazon Translate supports text translation between the languages listed in the following table.
				// The language code column uses ISO 639-1 two-digit language codes.
				// For a country variant of a language, the table follows the RFC 5646 format of appending a dash
				// followed by an ISO 3166 2-digit country code.
				// For example, the language code for the Mexican variant of Spanish is es-MX.
				// List of supported languages - https://docs.aws.amazon.com/translate/latest/dg/what-is-languages.html
				TargetLanguageCode: ptr(targetLanguage),

				// NOTE: replace auto with source language.
				// If you specify auto , you must send the TranslateText request in a region that supports Amazon Comprehend
				SourceLanguageCode: ptr("auto"),
				Text:               ptr(bufs[i].String()),
			})
		if err != nil {
			return nil, fmt.Errorf("translate text: %w", err)
		}

		// remove trailing newline.
		*translateOutput.TranslatedText = strings.TrimSuffix(*translateOutput.TranslatedText, "\n")
		translatedTexts = append(translatedTexts, strings.Split(*translateOutput.TranslatedText, "\n")...)
	}

	if len(messages.Messages) != len(translatedTexts) {
		return nil, errors.New("aws translated message count doesn't match input message count")
	}

	translatedMessages := model.Messages{
		Language: messages.Language,
		Original: messages.Original,
		Messages: make([]model.Message, 0, len(translatedTexts)),
	}

	for i := range translatedTexts {
		translatedMessages.Messages = append(translatedMessages.Messages, model.Message{
			ID:          messages.Messages[i].ID,
			PluralID:    messages.Messages[i].PluralID,
			Description: messages.Messages[i].Description,
			Message:     translatedTexts[i],
			Status:      model.MessageStatusFuzzy,
			Positions:   messages.Messages[i].Positions,
		})
	}

	return &translatedMessages, nil
}

// helpers

// ptr returns pointer to the passed in value.
func ptr[T any](v T) *T {
	return &v
}
