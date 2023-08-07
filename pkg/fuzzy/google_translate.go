package fuzzy

import (
	"context"
	"fmt"
	"io"

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

// googleTranslateRequestlimit limits the number of strings per translation request.
//
// google translate client: translate: rpc error:
// code = InvalidArgument
// desc = Number of strings in contents: 3830 exceeds the maximum limit of 1024
// error details:
//
//	name = BadRequest
//	field = contents
//	desc = Number of sub-requests should not be more than 1024
const googleTranslateRequestlimit = 1024

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

	n := len(messages.Messages)
	translatedMessages := model.Messages{
		Language: targetLanguage,
		Original: messages.Original,
		Messages: make([]model.Message, 0, n),
	}

	for low := 0; low < n; low += googleTranslateRequestlimit {
		high := low + googleTranslateRequestlimit

		if high > n {
			high = n
		}

		req := &translatepb.TranslateTextRequest{
			Parent:             parent(),
			SourceLanguageCode: messages.Language.String(),
			TargetLanguageCode: targetLanguage.String(),
			Contents:           make([]string, 0, high-low),
		}

		// Extract the strings to be translated by the Google Translate API.
		for i := low; i < high; i++ {
			req.Contents = append(req.Contents, messages.Messages[i].Message)
		}

		res, err := g.client.TranslateText(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("google translate client: translate text: %w", err)
		}

		for i, t := range res.Translations {
			m := messages.Messages[low+i]
			m.Message = t.TranslatedText
			m.Status = model.MessageStatusFuzzy

			translatedMessages.Messages = append(translatedMessages.Messages, m)
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
