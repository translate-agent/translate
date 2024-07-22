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

	mf2 "go.expect.digital/mf2/parse"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/expect"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// ---–––-------------- Tests------------------–––---

func Test_Translate(t *testing.T) {
	t.Parallel()

	targetLang := language.Latvian

	allTranslators(t, func(t *testing.T, translator Translator, subTest testutil.SubtestFn) {
		subTest("Multiple messages", func(ctx context.Context, t *testing.T) {
			input := rand.ModelTranslation(3, nil, rand.WithLanguage(language.English))

			output, err := translator.Translate(ctx, input, targetLang)
			expect.NoError(t, err)

			// Check the number of translated messages is the same as the number of input messages.
			expect.Equal(t, len(output.Messages), len(input.Messages))

			// Check the translated messages are not empty and are marked as fuzzy.
			for _, m := range output.Messages {
				require.NotEmpty(t, m.Message)
				expect.Equal(t, model.MessageStatusFuzzy, m.Status)

				_, err := mf2.Parse(m.Message)
				expect.NoError(t, err)
			}
		})
	})
}

// ------------------Helpers and init------------------

// translators is a map of all possible translation services, e.g. Google Translate, DeepL, etc.
var translators = map[string]Translator{}

// initAWSTranslate creates a new AWS Translate service and adds it to the translators map.
func initAWSTranslate(ctx context.Context) error {
	at, err := NewAWSTranslate(ctx, WithDefaultAWSClient(ctx))
	if err != nil {
		return fmt.Errorf("create new AWS Translate: %w", err)
	}

	translators["AWSTranslate"] = at

	return nil
}

// initGoogleTranslate creates a new Google Translate service and adds it to the translators map.
func initGoogleTranslate(ctx context.Context) (func() error, error) {
	gt, closer, err := NewGoogleTranslate(ctx, WithDefaultGoogleClient(ctx))
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
		log.Fatal(err)
	}

	// AWS Translate
	if err = initAWSTranslate(ctx); err != nil {
		log.Fatal(err)
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, subTest := testutil.Trace(t)

			f(t, translator, subTest)
		})
	}
}
