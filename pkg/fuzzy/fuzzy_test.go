package fuzzy

import (
	"context"
	"errors"
	"strings"
	"testing"

	"cloud.google.com/go/translate/apiv3/translatepb"
	awst "github.com/aws/aws-sdk-go-v2/service/translate"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// ---–––--------------Actual Tests------------------–––---

func Test_TranslateMock(t *testing.T) {
	// NOTE: AWS skipped for now until fuzzy translation is fixed.
	t.Parallel()

	allMocks(t, func(t *testing.T, mock Translator) {
		tests := []struct {
			expectedErr error
			translation *model.Translation
			name        string
		}{
			{
				name:        "One message",
				translation: randTranslation(1, language.Latvian),
			},
			{
				name:        "Multiple messages",
				translation: randTranslation(5, language.German),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				translation := tt.translation
				translation.Language = language.English // set original language
				translated, err := mock.Translate(context.Background(), translation, tt.translation.Language)

				if tt.expectedErr != nil {
					require.ErrorContains(t, err, tt.expectedErr.Error())
					return
				}

				require.NoError(t, err)

				// Check the that the translated translation have the correct language.
				require.Equal(t, tt.translation.Language, translated.Language)

				// Check that length matches.
				require.Len(t, translated.Messages, len(tt.translation.Messages))

				for i, m := range translated.Messages {
					require.Equal(t, model.MessageStatusFuzzy, m.Status)
					testutil.EqualMF2Message(t, tt.translation.Messages[i].Message, m.Message)

					// Reset the message to empty and fuzzy to original values, for the last check for side effects.
					translated.Messages[i].Status = tt.translation.Messages[i].Status
					translated.Messages[i].Message = tt.translation.Messages[i].Message
				}

				// Check the translated translation.messages are the same as the input messages. (Check for side effects)
				require.ElementsMatch(t, tt.translation.Messages, translated.Messages)
			})
		}
	})
}

// -------------------------Mocks------------------------------

// ---–––--------------Google Translate------------------–––---

// MockGoogleTranslateClient is a mock implementation of the Google Translate client.
type MockGoogleTranslateClient struct{}

// TranslateText mocks the Translate method of the Google Translate client.
func (m *MockGoogleTranslateClient) TranslateText(
	ctx context.Context,
	req *translatepb.TranslateTextRequest,
	opts ...gax.CallOption,
) (*translatepb.TranslateTextResponse, error) {
	// Mock the Bad request error for unsupported language.Afrikaans.
	if req.GetTargetLanguageCode() == language.Afrikaans.String() {
		return nil, errors.New("mock: bad request: unsupported language")
	}

	res := &translatepb.TranslateTextResponse{
		Translations: make([]*translatepb.Translation, 0, len(req.GetContents())),
	}

	for _, v := range req.GetContents() {
		res.Translations = append(res.GetTranslations(), &translatepb.Translation{TranslatedText: v})
	}

	return res, nil
}

func (m *MockGoogleTranslateClient) Close() error { return nil }

// ---–––--------------AWS Translate------------------–––---

// MockAWSTranslateClient is a mock implementation of the AWS Translate client.
type MockAWSTranslateClient struct{}

// Translate mocks the TranslateText method of the AWS Translate client.
func (m *MockAWSTranslateClient) TranslateText(
	ctx context.Context,
	params *awst.TranslateTextInput,
	optFns ...func(*awst.Options),
) (*awst.TranslateTextOutput, error) {
	// Mock the Bad request error for unsupported language.Und.
	if *params.SourceLanguageCode == language.Und.String() ||
		*params.TargetLanguageCode == language.Und.String() {
		return nil, errors.New("mock: bad request: Unsupported language pair: zh to und")
	}

	// remove trailing newline character.
	*params.Text = strings.TrimSuffix(*params.Text, "\n")
	translations := strings.Split(*params.Text, "\n")

	for i := range translations {
		translations[i] = gofakeit.SentenceSimple()
	}

	return &awst.TranslateTextOutput{
		SourceLanguageCode: params.SourceLanguageCode,
		TargetLanguageCode: params.TargetLanguageCode,
		TranslatedText:     ptr(strings.Join(translations, "\n")),
	}, nil
}

// -----------------------Helpers and init----------------------------

var mockTranslators map[string]Translator

func init() {
	mockTranslators = make(map[string]Translator, len(SupportedServices))

	// TODO
	// at, _ := NewAWSTranslate(
	// 	context.Background(),
	// 	WithAWSClient(&MockAWSTranslateClient{}),
	// )

	// mockTranslators["AWSTranslate"] = at

	// Google Translate
	gt, _, _ := NewGoogleTranslate(
		context.Background(),
		WithGoogleClient(&MockGoogleTranslateClient{}),
	)

	mockTranslators["GoogleTranslate"] = gt
}

// allMocks runs a test function f for each mocked translate service that is defined in the mockTranslators map.
func allMocks(t *testing.T, f func(t *testing.T, mock Translator)) {
	for name, mock := range mockTranslators {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f(t, mock)
		})
	}
}

// randTranslation returns a random translation model with the given count of messages and source language.
// The messages will not be fuzzy.
func randTranslation(msgCount uint, srcLang language.Tag) *model.Translation {
	msgOpts := []rand.ModelMessageOption{rand.WithStatus(model.MessageStatusUntranslated)}
	translationOpts := []rand.ModelTranslationOption{rand.WithLanguage(srcLang)}

	return rand.ModelTranslation(msgCount, msgOpts, translationOpts...)
}
