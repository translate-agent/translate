package fuzzy

import (
	"context"
	"testing"

	"cloud.google.com/go/translate/apiv3/translatepb"
	awst "github.com/aws/aws-sdk-go-v2/service/translate"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// ---–––--------------Actual Tests------------------–––---

func Test_TranslateMock(t *testing.T) {
	t.Parallel()

	targetLang := language.Latvian

	allMocks(t, func(t *testing.T, mock Translator) {
		tests := []struct {
			input *model.Translation
			name  string
		}{
			{
				name:  "One message",
				input: rand.ModelTranslation(1, nil, rand.WithLanguage(language.English)),
			},
			{
				name:  "Multiple messages",
				input: rand.ModelTranslation(3, nil, rand.WithLanguage(language.English)),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				output, err := mock.Translate(context.Background(), tt.input, targetLang)
				require.NoError(t, err)

				// Check the that the translated translation have the correct language.
				require.Equal(t, targetLang, output.Language)

				// Check that length matches.
				require.Len(t, output.Messages, len(tt.input.Messages))

				for i, m := range output.Messages {
					require.Equal(t, model.MessageStatusFuzzy, m.Status)
					testutil.EqualMF2Message(t, tt.input.Messages[i].Message, m.Message)

					// Reset the message to empty and fuzzy to original values, for the last check for side effects.
					output.Messages[i].Status = tt.input.Messages[i].Status
					output.Messages[i].Message = tt.input.Messages[i].Message
				}

				// Check the translated translation.messages are the same as the input messages. (Check for side effects)
				require.ElementsMatch(t, tt.input.Messages, output.Messages)
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
	translations := make([]*translatepb.Translation, 0, len(req.GetContents()))
	for _, v := range req.GetContents() {
		translations = append(translations, &translatepb.Translation{TranslatedText: v})
	}

	return &translatepb.TranslateTextResponse{Translations: translations}, nil
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
	return &awst.TranslateTextOutput{
		SourceLanguageCode: params.SourceLanguageCode,
		TargetLanguageCode: params.TargetLanguageCode,
		TranslatedText:     params.Text,
	}, nil
}

// -----------------------Helpers and init----------------------------

var mockTranslators map[string]Translator

func init() {
	mockTranslators = make(map[string]Translator, len(SupportedServices))

	at, _ := NewAWSTranslate(
		context.Background(),
		WithAWSClient(&MockAWSTranslateClient{}),
	)

	mockTranslators["AWSTranslate"] = at

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
