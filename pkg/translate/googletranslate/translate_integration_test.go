//go:build integration

package googletranslate

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/translate"
	"golang.org/x/text/language"
)

var translateService translate.TranslationService

func TestMain(m *testing.M) {
	code := testMain(m)
	os.Exit(code)
}

func testMain(m *testing.M) (code int) {
	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	apiKey := viper.GetString("googletranslate.api.key")
	if apiKey == "" {
		log.Println("no Google Translate API key provided, skipping integration tests")
		return 0
	}

	var (
		err    error
		closer func() error
	)

	translateService, closer, err = NewGoogleTranslate(ctx, WithDefaultClient(ctx, apiKey))
	if err != nil {
		log.Panicf("create new google translate service: %v", err)
	}

	// Try to close the Google Translate service after the tests have finished.
	defer func() {
		if err := closer(); err != nil {
			log.Printf("close google translate service: %v", err)
		}
	}()

	return m.Run()
}

// Test_GoogleTranslate tests the Translate method of the Google Translate service using a real client.
func Test_GoogleTranslate(t *testing.T) {
	t.Parallel()

	_, subTest := testutil.Trace(t)

	tests := []struct {
		messages   *model.Messages
		targetLang language.Tag
		name       string
	}{
		{
			name:       "One message",
			messages:   randMessages(1, language.English),
			targetLang: language.Latvian,
		},
		{
			name:       "Undefined language messages",
			messages:   randMessages(7, language.Und),
			targetLang: language.Latvian,
		},
	}

	for _, tt := range tests {
		tt := tt
		subTest(tt.name, func(ctx context.Context, t *testing.T) {
			translatedMsgs, err := translateService.Translate(ctx, tt.messages, tt.targetLang)
			require.NoError(t, err)

			// Check the number of translated messages is the same as the number of input messages.
			require.Len(t, translatedMsgs.Messages, len(tt.messages.Messages))

			// Check the translated messages are not empty and are marked as fuzzy.
			for _, m := range translatedMsgs.Messages {
				require.NotEmpty(t, m.Message)
				require.True(t, m.Fuzzy)
			}
		})
	}
}
