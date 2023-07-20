//go:build integration

package fuzzy

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	awstranslate "go.expect.digital/translate/pkg/fuzzy/aws"
	googletranslate "go.expect.digital/translate/pkg/fuzzy/google"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

// ---–––--------------Actual Tests------------------–––---

func Test_Translate(t *testing.T) {
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
	"AWSTranslate":    nil,
}

// initAWSTranslate creates a new AWS Translate service and adds it to the translators map.
func initAWSTranslate(ctx context.Context) error {
	at, err := awstranslate.NewAWSTranslate(ctx, awstranslate.WithDefaultClient(ctx))
	if err != nil {
		return fmt.Errorf("create new AWS Translate: %w", err)
	}

	translators["AWSTranslate"] = at

	return nil
}

// initGoogleTranslate creates a new Google Translate service and adds it to the translators map.
func initGoogleTranslate(ctx context.Context) (func() error, error) {
	gt, closer, err := googletranslate.NewGoogleTranslate(ctx, googletranslate.WithDefaultClient(ctx))
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
	gtCloser, err := initGoogleTranslate(ctx)
	if err != nil {
		// If the Google Translate API credentials are not set, skip the Google Translate tests.
		if strings.Contains(err.Error(), "could not find default credentials") {
			log.Println("Google Translate API credentials are not set. Skipping Google Translate tests.")
		} else {
			// All other errors are fatal.
			log.Fatal(err)
		}
	}

	// AWS Translate
	if err = initAWSTranslate(ctx); err != nil {
		// If the AWS Translate access or secret key is not set, skip the AWS Translate tests.
		if strings.Contains(err.Error(), "key is not set") {
			log.Printf("%s. Skipping AWS Translate tests.", err)
		} else {
			// All other errors are fatal.
			log.Fatal(err)
		}
	}

	// Close all connections

	// Close the Google Translate client.
	if gtCloser != nil {
		defer func() {
			if err := gtCloser(); err != nil {
				log.Printf("close Google Translate: %v", err)
			}
		}()
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
