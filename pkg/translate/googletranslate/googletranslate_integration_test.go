//go:build integration

package googletranslate

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	googleTranslate "cloud.google.com/go/translate"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

var (
	client *googleTranslate.Client
	apiKey string
)

func TestMain(m *testing.M) {
	code := testMain(m)
	os.Exit(code)
}

func testMain(m *testing.M) (code int) {
	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	apiKey = viper.GetString("googletranslate.api.key")

	var err error
	client, err = googleTranslate.NewClient(ctx, option.WithAPIKey(apiKey))
	// Ignore error if the error is about not finding default credentials.
	// In that case, integration tests will be skipped.
	if err != nil && !strings.Contains(err.Error(), "could not find default credentials") {
		log.Panicf("create new google translate client: %v", err)
	}

	defer client.Close()

	return m.Run()
}

// Test_GoogleTranslate tests the Translate method of the Google Translate service using a real client.
func Test_GoogleTranslate(t *testing.T) {
	t.Parallel()

	if apiKey == "" {
		t.Skip("Google Translate API key not set")
	}

	ctx, subTest := testutil.Trace(t)

	translateService, err := NewGoogleTranslate(ctx, client)
	require.NoError(t, err)

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
