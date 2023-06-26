//go:build integration

package translatetest

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/translate"
	"go.expect.digital/translate/pkg/translate/googletranslate"

	"go.expect.digital/translate/pkg/testutil"
)

// alias for the translate service interface.
type service translate.TranslationService

// translators is a map of all possible translation services, e.g. Google Translate, DeepL, etc.
var translators = map[string]service{
	"GoogleTranslate": nil,
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

func TestMain(m *testing.M) {
	code := testMain(m)

	os.Exit(code)
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
		// If the Google Translate API key is not set, skip the Google Translate tests.
		if strings.Contains(err.Error(), "api key is not set") {
			log.Println("Google Translate API key is not set. Skipping Google Translate tests.")
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

// allServices runs a test for each translate service that is defined in the translators map.
func allServices(t *testing.T, f func(t *testing.T, service service, subtest testutil.SubtestFn)) {
	for name, service := range translators {
		name, service := name, service
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if service == nil {
				t.Skipf("'%s' is not initialized", name)
			}

			_, subTest := testutil.Trace(t)

			f(t, service, subTest)
		})
	}
}
