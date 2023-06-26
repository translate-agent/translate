//go:build integration

package translatetest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

// Test_Translate tests if the integration with the provided translation service is working.
// Translate logic is tested in each translation service's package in translate_test.go.
func Test_Translate(t *testing.T) {
	t.Parallel()

	allServices(t, func(t *testing.T, service service, subTest testutil.SubtestFn) {
		subTest("One message", func(ctx context.Context, t *testing.T) {
			messages := RandMessages(1, language.English)

			translatedMsgs, err := service.Translate(ctx, messages, language.Latvian)
			require.NoError(t, err)

			// Check the number of translated messages is the same as the number of input messages.
			require.Len(t, translatedMsgs.Messages, len(messages.Messages))

			// Check the translated messages are not empty and are marked as fuzzy.
			for _, m := range translatedMsgs.Messages {
				require.NotEmpty(t, m.Message)
				require.True(t, m.Fuzzy)
			}
		})
	})
}
