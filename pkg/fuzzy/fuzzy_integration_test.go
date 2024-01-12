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
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

// ---–––--------------Actual Tests------------------–––---

func Test_Translate(t *testing.T) {
	t.Parallel()

	allTranslators(t, func(t *testing.T, translator Translator, subTest testutil.SubtestFn) {
		subTest("Multiple translations", func(ctx context.Context, t *testing.T) {
			translation := translationWithMF2Messages() // TODO: create translation with randomized MF2 messages.
			translation.Language = language.English     // set original language
			translated, err := translator.Translate(ctx, translation, language.Latvian)
			require.NoError(t, err)

			// Check the number of translated messages is the same as the number of input messages.
			require.Len(t, translated.Messages, len(translation.Messages))

			// Check the translated messages are not empty and are marked as fuzzy.
			for _, m := range translated.Messages {
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
	"AWSTranslate":    nil, // TODO
}

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
		name, translator := name, translator
		t.Run(name, func(t *testing.T) {
			if name == "AWSTranslate" { // TODO
				t.Skip()
			}

			t.Parallel()

			if translator == nil {
				t.Skipf("'%s' is not initialized", name)
			}

			_, subTest := testutil.Trace(t)

			f(t, translator, subTest)
		})
	}
}

// TODO: Temporary  implementation, remove after implementing randomized MF2 messages.
func translationWithMF2Messages() *model.Translation {
	translation := model.Translation{
		Original: true,
		Messages: []model.Message{
			{
				ID:      "1",
				Message: "Hello, world!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "2",
				Message: "Hello, \\{World!\\}",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "3",
				Message: "{ $variable } Hello, World!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "4",
				Message: "Hello, World! { $variable }",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "5",
				Message: "Hello, { $variable :function }  World!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "6",
				Message: "Hello, { $variable :function option1 = -3.14 ns:option2 = |value2| option3 = $variable2 } World!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "7",
				Message: "Hello, { |literal| }  World!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "8",
				Message: "Hello, { |name| :function ns1:option1 = -1 ns2:option2 = 1 option3 = |value3| } World!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "9",
				Message: "Hello, { |name| :function } World!",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "10",
				Message: "{{Hello, { |literal| } World!}}",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "11",
				Message: ".local $var={2} {{Hello world}}",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID:      "12",
				Message: ".local $var = { $anotherVar } {{Hello { $var } world}}",
				Status:  model.MessageStatusUntranslated,
			},
			{
				ID: "13",
				Message: `.local $var = { :ns1:function opt1 = 1 opt2 = |val2| }
			 .local $var = { 2 } {{Hello { $var :ns2:function2 } world}}`,
				Status: model.MessageStatusUntranslated,
			},
			{
				ID: "14",
				Message: `.match { $variable :number }
							1 {{Hello { $variable } world}}
							* {{Hello { $variable } worlds}}`,
				Status: model.MessageStatusUntranslated,
			},
			{
				ID: "15",
				Message: `.local $var1 = { male }
							.local $var2 = { |female| }
							.match { :gender }
							male {{Hello sir!}}
							|female| {{Hello madam!}}
							* {{Hello { $var1 } or { $var2 }!}}`,
				Status: model.MessageStatusUntranslated,
			},
			//nolint:dupword
			{
				ID: "16",
				Message: `.match { $var1 } { $var2 }
						yes yes {{Hello beautiful world!}}
						yes no {{Hello beautiful!}}
						no yes {{Hello world!}}
						no no {{Hello!}}`,
				Status: model.MessageStatusUntranslated,
			},
		},
	}

	return &translation
}
