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
	googleTranslateRequestLimit = 1024 //nolint:unused

	// googleTranslateCodePointsLimit limits the number of Unicode codepoints per single translation request.
	googleTranslateCodePointsLimit = 30_000 //nolint:unused
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
			option.WithGRPCDialOption(grpc.WithStatsHandler(otelgrpc.NewClientHandler())),
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
	translation *model.Translation,
	targetLanguage language.Tag,
) (*model.Translation, error) {
	if translation == nil {
		return nil, nil //nolint:nilnil
	}

	// TODO: Implement translation of message texts.

	// NOTE: temporary fix to avoid failing tests.
	if len(translation.Messages) == 0 {
		return &model.Translation{Language: targetLanguage, Original: translation.Original}, nil
	}

	translated := &model.Translation{
		Language: targetLanguage,
		Original: translation.Original,
		Messages: make([]model.Message, len(translation.Messages)),
	}

	for i := range translation.Messages {
		translated.Messages = append(translated.Messages, translation.Messages[i])
		translated.Messages[i].Status = model.MessageStatusFuzzy
	}

	// NOTE: Previous implementation of the translation of message texts, for reference!

	// // Extract translatable text from translation.
	// asts, err := mf.ParseTranslation(translation)
	// if err != nil {
	// 	return nil, fmt.Errorf("google translate: parse translation messages: %w", err)
	// }

	// textNodes := mf.GetTextNodes(asts)
	// texts := textNodes.GetTexts()

	// // Split text from translation into batches to avoid exceeding
	// // googleTranslateRequestLimit or googleTranslateCodePointsLimit.
	// batches := textToBatches(texts, googleTranslateRequestLimit)
	// translatedTexts := make([]string, 0, len(texts))

	// // Translate text batches using Google Translate client.
	// for i := range batches {
	// 	res, err := g.client.TranslateText(ctx, &translatepb.TranslateTextRequest{
	// 		Parent:             parent(),
	// 		SourceLanguageCode: translation.Language.String(),
	// 		TargetLanguageCode: targetLanguage.String(),
	// 		Contents:           batches[i],
	// 	})
	// 	if err != nil {
	// 		return nil, fmt.Errorf("google translate client: translate text #%d from batch: %w", i, err)
	// 	}

	// 	for i := range res.GetTranslations() {
	// 		translatedTexts = append(translatedTexts, res.GetTranslations()[i].GetTranslatedText())
	// 	}
	// }

	// // Overwrite text nodes in ASTs to include newly translated text.
	// if err := textNodes.OverwriteTexts(translatedTexts); err != nil {
	// 	return nil, fmt.Errorf("google translate: overwrite text nodes in ASTs: %w", err)
	// }

	// // create translation with newly translated messages.
	// translated := model.Translation{
	// 	Language: targetLanguage,
	// 	Original: translation.Original,
	// 	Messages: make([]model.Message, len(translation.Messages)),
	// }

	// for i := range asts {
	// 	b, err := asts[i].MarshalText()
	// 	if err != nil {
	// 		return nil,
	// 			fmt.Errorf("google translate: marshal text from AST for message ID '%s': %w", translation.Messages[i].ID, err)
	// 	}

	// 	m := translation.Messages[i]
	// 	m.Message = string(b)
	// 	m.Status = model.MessageStatusFuzzy
	// 	translated.Messages[i] = m
	// }

	// return &translated, nil

	return translated, nil
}

// helpers

// parent returns path to Google project and location.
func parent() string {
	projectID := viper.GetString("other.google.project_id")
	location := viper.GetString("other.google.location")

	return fmt.Sprintf("projects/%s/locations/%s", projectID, location)
}

// TODO: Remove if not used.
// textToBatches splits text into batches with predefined maximum amount of elements.
func textToBatches(text []string, batchLimit int) [][]string { //nolint:unused
	var codePointsInBatch int

	batch := make([]string, 0, batchLimit)
	batches := make([][]string, 0, 1)

	for _, text := range text {
		codePointsInMsg := utf8.RuneCountInString(text)

		if len(batch) == googleTranslateRequestLimit || codePointsInBatch+codePointsInMsg > googleTranslateCodePointsLimit {
			batches = append(batches, batch)
			batch = make([]string, 0, googleTranslateRequestLimit)

			codePointsInBatch = 0
		}

		batch = append(batch, text)
		codePointsInBatch += codePointsInMsg
	}

	batches = append(batches, batch)

	return batches
}
