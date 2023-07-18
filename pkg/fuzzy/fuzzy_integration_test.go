//go:build integration

package fuzzy

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	translatev3 "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// ---–––--------------Actual Tests------------------–––---

func Test_TranslateText(t *testing.T) {
	t.Parallel()

	allTranslators(t, func(t *testing.T, translator Translator, subTest testutil.SubtestFn) {
		subTest("Multiple messages", func(ctx context.Context, t *testing.T) {
			messages := randMessages(3, language.Latvian)

			translatedMsgs, err := translator.Translate(ctx, messages)
			require.NoError(t, err)

			// Check the number of translated messages is the same as the number of input messages.
			require.Len(t, translatedMsgs.Messages, len(messages.Messages))

			// Check the translated messages are not empty and are marked as fuzzy.
			for _, m := range translatedMsgs.Messages {
				require.NotEmpty(t, m.Message)
				require.Equal(t, model.MessageStatusFuzzy, m.Status)
			}
		})
	})
}

// ------------------Helpers and init------------------

// translators is a map of all possible translation services, e.g. Google Translate, DeepL, etc.
var translators = map[string]Translator{
	"GoogleTranslate": nil,
}

// initGoogleTranslate creates a new Google Translate service and adds it to the translators map.
func initGoogleTranslate(ctx context.Context) (func() error, error) {
	gt, closer, err := NewGoogleTranslate(ctx, WithDefaultClient(ctx))
	if err != nil {
		return nil, fmt.Errorf("create new Google Translate: %w", err)
	}

	translators["GoogleTranslate"] = gt

	return closer, nil
}

func testMain(m *testing.M) int {
	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	// Initialize all translation services.

	// Google Translate
	client, err := translatev3.NewTranslationClient(ctx, option.WithAPIKey(viper.GetString("GOOGLE_TRANSLATE_API_KEY")))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Construct a request.
	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/%s", viper.GetString("GOOGLE_PROJECT_ID"), viper.GetString("GOOGLE_LOCATION")),
		Contents:           []string{"hello, world!"},
		MimeType:           "text/plain", // Mime types: "text/plain", "text/html"
		SourceLanguageCode: "en-US",
		TargetLanguageCode: "fr-FR",
	}

	response, err := client.TranslateText(ctx, req)
	if err != nil {
		log.Fatalf("Failed to translate text: %v", err)
	}

	for _, translation := range response.Translations {
		fmt.Printf("Translated text: %v\n", translation.GetTranslatedText())
	}

	return m.Run()
}

func TestMain(m *testing.M) {
	code := testMain(m)

	os.Exit(code)
}

// allTranslators runs a test for each translate service that is defined in the translators map.
func allTranslators(t *testing.T, f func(t *testing.T, translator Translator, subtest testutil.SubtestFn)) {
	for name, translator := range translators {
		name, translator := name, translator
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if translator == nil {
				t.Skipf("'%s' is not initialized", name)
			}

			_, subTest := testutil.Trace(t)

			f(t, translator, subTest)
		})
	}
}
