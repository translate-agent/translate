package fuzzy

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"unicode/utf8"

	ast "go.expect.digital/mf2/parse"

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

	// NOTE: temporary fix to avoid failing tests.
	if len(translation.Messages) == 0 {
		return &model.Translation{Language: targetLanguage, Original: translation.Original}, nil
	}

	// Retrieve all translatable text from translation
	texts, err := getTexts(translation)
	if err != nil {
		return nil, fmt.Errorf("google translate: get texts: %w", err)
	}

	// Split text from translation into batches to avoid exceeding
	// googleTranslateRequestLimit or googleTranslateCodePointsLimit.
	batches := textToBatches(texts, googleTranslateRequestLimit)
	translatedTexts := make([]string, 0, len(texts))

	// Translate text batches using Google Translate client.
	for i := range batches {
		res, err := g.client.TranslateText(ctx, &translatepb.TranslateTextRequest{ //nolint:govet
			Parent:             parent(),
			SourceLanguageCode: translation.Language.String(),
			TargetLanguageCode: targetLanguage.String(),
			Contents:           batches[i],
		})
		if err != nil {
			return nil, fmt.Errorf("google translate client: translate text #%d from batch: %w", i, err)
		}

		for i := range res.GetTranslations() {
			translatedTexts = append(translatedTexts, res.GetTranslations()[i].GetTranslatedText())
		}
	}

	// build translation with new translated text
	translated, err := buildTranslated(translation, translatedTexts, targetLanguage)
	if err != nil {
		return nil, fmt.Errorf("google translate: build translated: %w", err)
	}

	return translated, nil
}

// helpers

// parent returns path to Google project and location.
func parent() string {
	projectID := viper.GetString("other.google.project_id")
	location := viper.GetString("other.google.location")

	return fmt.Sprintf("projects/%s/locations/%s", projectID, location)
}

// textToBatches splits text into batches with predefined maximum amount of elements.
func textToBatches(text []string, batchLimit int) [][]string {
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

func getTexts(translation *model.Translation) ([]string, error) {
	texts := make([]string, 0, len(translation.Messages))

	for i := range translation.Messages {
		messageAST, err := ast.Parse(translation.Messages[i].Message)
		if err != nil {
			return nil, fmt.Errorf("parse mf2 message: %w", err)
		}

		switch v := messageAST.Message.(type) {
		case ast.SimpleMessage:
			texts = append(texts, printPattern(v.Patterns))
		case ast.ComplexMessage:
			switch v := v.ComplexBody.(type) {
			case ast.Matcher:
				for _, variant := range v.Variants {
					texts = append(texts, printPattern(variant.QuotedPattern.Patterns))
				}
			case ast.QuotedPattern:
				texts = append(texts, printPattern(v.Patterns))
			}
		}
	}

	return texts, nil
}

// printPattern ranges over pattern appending TextPatterns to a string
// when a placeholderPattern is encountered - {$d} is appended, returns resulting string.
// input:
//
//	[]ast.Patterns{
//		TextPattern("Hello"),
//		PlaceholderPattern{ Expression: LiteralExpression{Literal: QuotedLiteral("name")}},
//		}
//
// output:
// "Hello {$0}"
// .
func printPattern(pattern []ast.Pattern) string {
	var (
		phIndex int
		text    string
	)

	for i := range pattern {
		switch v := pattern[i].(type) {
		case ast.TextPattern:
			text += string(v)
		case ast.PlaceholderPattern:
			text += fmt.Sprintf("{$%d}", phIndex)
			phIndex++
		}
	}

	return text
}

func buildTranslated(translation *model.Translation, translatedTexts []string, targetLanguage language.Tag,
) (*model.Translation, error) {
	translated := &model.Translation{
		Language: targetLanguage,
		Original: translation.Original,
		Messages: make([]model.Message, len(translation.Messages)),
	}

	var textIndex int

	for i := range translation.Messages {
		messageAST, err := ast.Parse(translation.Messages[i].Message)
		if err != nil {
			return nil, fmt.Errorf("parse mf2 message: %w", err)
		}

		switch message := messageAST.Message.(type) {
		case ast.SimpleMessage:
			message.Patterns, err = buildTranslatedPattern(translatedTexts[textIndex], message.Patterns)
			if err != nil {
				return nil, fmt.Errorf("build translated pattern for simple message: %w", err)
			}

			textIndex++

			// rewrite AST
			messageAST.Message = message
		case ast.ComplexMessage:
			switch complexBody := message.ComplexBody.(type) {
			case ast.Matcher:
				for i := range complexBody.Variants {
					complexBody.Variants[i].QuotedPattern.Patterns, err = buildTranslatedPattern(
						translatedTexts[textIndex], complexBody.Variants[i].QuotedPattern.Patterns)
					if err != nil {
						return nil, fmt.Errorf("build translated pattern for matcher variant: %w", err)
					}

					textIndex++
				}

				// rewrite AST
				message.ComplexBody = complexBody
				messageAST.Message = message
			case ast.QuotedPattern:
				complexBody.Patterns, err = buildTranslatedPattern(translatedTexts[textIndex], complexBody.Patterns)
				if err != nil {
					return nil, fmt.Errorf("build translated pattern for quoted pattern: %w", err)
				}

				textIndex++

				// rewrite AST
				message.ComplexBody = complexBody
				messageAST.Message = message
			}
		}

		translated.Messages[i] = model.Message{
			ID:          translation.Messages[i].ID,
			PluralID:    translation.Messages[i].PluralID,
			Message:     messageAST.String(),
			Description: translation.Messages[i].Description,
			Positions:   translation.Messages[i].Positions,
			Status:      model.MessageStatusFuzzy,
		}
	}

	return translated, nil
}

func buildTranslatedPattern(translatedText string, patterns []ast.Pattern) ([]ast.Pattern, error) {
	re := regexp.MustCompile(`\{\$(\d+)\}`)
	textParts := splitText(translatedText)
	placeholders := getPlaceholders(patterns)
	translatedPatterns := make([]ast.Pattern, 0, len(patterns))

	for i, v := range textParts {
		if re.MatchString(textParts[i]) { // placeholder
			placeholderIndex, err := strconv.ParseInt(v[2:len(v)-1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse placeholder index: %w", err)
			}

			translatedPatterns = append(translatedPatterns, placeholders[placeholderIndex])
		} else { // text
			translatedPatterns = append(translatedPatterns, ast.TextPattern(textParts[i]))
		}
	}

	return translatedPatterns, nil
}

func getPlaceholders(patterns []ast.Pattern) []ast.PlaceholderPattern {
	placeholders := make([]ast.PlaceholderPattern, 0, len(patterns))

	for i := range patterns {
		switch v := patterns[i].(type) {
		default:
			// noop
		case ast.PlaceholderPattern:
			placeholders = append(placeholders, v)
		}
	}

	return placeholders
}

// splitText splits string into all substrings separated by {$d}
// and returns a slice of the substrings including separators.
// input:
// "Hello {$0} {$1}! Welcome to {$2}."
// output:
// []string{"Hello ", "{$0}", " ", "{$1}" "! Welcome to ", "{$2}"}
// .
func splitText(s string) []string {
	var startIndex int

	parts := make([]string, 0, 1)
	indices := regexp.MustCompile(`\{\$(\d+)\}`).FindAllStringIndex(s, -1)

	for _, indexPair := range indices {
		if s[startIndex:indexPair[0]] != "" {
			parts = append(parts, s[startIndex:indexPair[0]])
		} else if s[indexPair[0]:indexPair[1]] != "" {
			parts = append(parts, s[indexPair[0]:indexPair[1]])
		}

		startIndex = indexPair[1]
	}

	if s[startIndex:] != "" {
		parts = append(parts, s[startIndex:])
	}

	return parts
}
