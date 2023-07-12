package awstranslate

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/translate"
	"github.com/aws/aws-sdk-go-v2/service/translate/types"

	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/model"
)

// --------------------Definitions--------------------

type awsClient interface {
	TranslateDocument(
		ctx context.Context,
		params *translate.TranslateDocumentInput,
		optFns ...func(*translate.Options),
	) (*translate.TranslateDocumentOutput, error)
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

		// Create a new AWS SDK config
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithHTTPClient(http.DefaultClient),
			config.WithRegion(viper.GetString("other.aws_translate.region")),
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
	_, err := tr.client.TranslateDocument(ctx, &translate.TranslateDocumentInput{
		SourceLanguageCode: ptr("en"),
		TargetLanguageCode: ptr("lv"),
		Document: &types.Document{
			Content:     []byte("Hello World!"),
			ContentType: ptr("text/plain"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("aws translate client: ping aws translate: %w", err)
	}

	return tr, nil
}

// Maximum text size limit accepted by the AWS Translate API - 100KB.
const inputTextSizeLimit = 100 * 1024

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
		if buf.Cap()+len(m.Message) > inputTextSizeLimit {
			bufs = append(bufs, buf)
			buf.Reset()
		}

		fmt.Fprintln(&buf, m.Message)
	}

	bufs = append(bufs, buf)

	translatedTexts := make([]string, 0, len(messages.Messages))

	for i := range bufs {
		translateOutput, err := tr.client.TranslateDocument(ctx,
			&translate.TranslateDocumentInput{
				SourceLanguageCode: ptr("en"),
				TargetLanguageCode: ptr("lv"),
				Document: &types.Document{
					Content:     bufs[i].Bytes(),
					ContentType: ptr("text/plain"),
				},
			})
		if err != nil {
			return nil, fmt.Errorf("make translate content request: %w", err)
		}

		scanner := bufio.NewScanner(
			bytes.NewBuffer(translateOutput.TranslatedDocument.Content),
		)

		for scanner.Scan() {
			line := scanner.Text()
			translatedTexts = append(translatedTexts, line)
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan translated texts from buffer: %w", err)
		}
	}

	if len(messages.Messages) != len(translatedTexts) {
		return nil, errors.New("translated texts count doesn't match input message count")
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
