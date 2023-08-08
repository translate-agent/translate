package fuzzy

import (
	"context"
	"fmt"
	"io"
	"unicode/utf8"

	translate "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// List of the google request limits based on TranslateTextRequest.contents description:
// https://github.com/googleapis/googleapis/blob/master/google/cloud/translate/v3/translation_service.proto
const (
	// googleTranslateRequestLimit limits the number of strings per translation request.
	googleTranslateRequestLimit = 1024

	// googleTranslateCodePointsLimit limits the number of Unicode codepoints per single translation request.
	googleTranslateCodePointsLimit = 30_000
)

// --------------------Definitions--------------------

// Interface that defines some of the methods of the Google Translate client.
// This interface helps to mock the Google Translate client in unit tests.
// https://pkg.go.dev/cloud.google.com/go/translate#Client
type googleClient interface {
	TranslateText(
		ctx context.Context,
		req *translatepb.TranslateTextRequest,
		opts ...gax.CallOption,
	) (*translatepb.TranslateTextResponse, error)
	io.Closer
}

// GoogleTranslate implements the Translator interface.
type GoogleTranslate struct {
	client googleClient
}

type GoogleTranslateOption func(*GoogleTranslate) error

// WithGoogleClient sets the Google Translate client.
func WithGoogleClient(c googleClient) GoogleTranslateOption {
	return func(g *GoogleTranslate) error {
		g.client = c
		return nil
	}
}

// WithDefaultGoogleClient creates a new Google Translate client with the API key from the viper.
func WithDefaultGoogleClient(ctx context.Context) GoogleTranslateOption {
	return func(g *GoogleTranslate) error {
		var err error

		// Create new Google Cloud service transport with the base of OpenTelemetry HTTP transport.
		g.client, err = translate.NewTranslationClient(ctx,
			option.WithCredentialsFile(viper.GetString("other.google.account_key")),
			option.WithGRPCDialOption(grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor())),
			option.WithGRPCDialOption(grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor())),
		)
		if err != nil {
			return fmt.Errorf("init Google translate client: %w", err)
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
	req := &translatepb.TranslateTextRequest{
		Parent:             parent(),
		TargetLanguageCode: language.Latvian.String(),
		Contents:           []string{"Hello World!"},
	}

	_, err = gt.client.TranslateText(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("google translate client: ping: %w", err)
	}

	return gt, gt.client.Close, nil
}

// --------------------Methods--------------------

func (g *GoogleTranslate) Translate(
	ctx context.Context,
	messages *model.Messages,
	targetLanguage language.Tag,
) (*model.Messages, error) {
	if messages == nil {
		return nil, nil
	}

	if len(messages.Messages) == 0 {
		return &model.Messages{Language: messages.Language, Original: messages.Original}, nil
	}

	translatedMessages := model.Messages{
		Language: targetLanguage,
		Original: messages.Original,
		Messages: make([]model.Message, 0, len(messages.Messages)),
	}

	// Split text from messages into batches to avoid exceeding
	// googleTranslateRequestLimit or googleTranslateCodePointsLimit.

	var codePointsInBatch int

	batch := make([]string, 0)
	batches := make([][]string, 0)

	for _, msgs := range messages.Messages {
		codePointsInMsg := utf8.RuneCountInString(msgs.Message)

		if len(batch) == googleTranslateRequestLimit || codePointsInBatch+codePointsInMsg > googleTranslateCodePointsLimit {
			batches = append(batches, batch)
			batch = make([]string, 0)

			codePointsInBatch = 0
		}

		batch = append(batch, msgs.Message)
		codePointsInBatch += codePointsInMsg
	}

	batches = append(batches, batch)

	// Translate text batches using Google Translate client.

	var msgsIndex int

	for i := range batches {
		res, err := g.client.TranslateText(ctx, &translatepb.TranslateTextRequest{
			Parent:             parent(),
			SourceLanguageCode: messages.Language.String(),
			TargetLanguageCode: targetLanguage.String(),
			Contents:           batches[i],
		})
		if err != nil {
			return nil, fmt.Errorf("google translate client: translate text from batch at index %d: %w", i, err)
		}

		for _, t := range res.Translations {
			m := messages.Messages[msgsIndex]
			m.Message = t.TranslatedText
			m.Status = model.MessageStatusFuzzy

			translatedMessages.Messages = append(translatedMessages.Messages, m)

			msgsIndex++
		}
	}

	return &translatedMessages, nil
}

// parent returns path to Google project and location.
func parent() string {
	projectId := viper.GetString("other.google.project_id")
	location := viper.GetString("other.google.location")

	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}
