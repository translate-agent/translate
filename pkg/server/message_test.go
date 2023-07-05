package server

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo/badgerdb"
	"go.expect.digital/translate/pkg/repo/common"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

var translateSrv *TranslateServiceServer

type mockTranslator struct{}

func (m *mockTranslator) Translate(ctx context.Context, messages *model.Messages, targetLang language.Tag) (*model.Messages, error) {
	return messages, nil
}

func TestMain(m *testing.M) {
	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	repository, err := badgerdb.NewRepo(badgerdb.WithDefaultDB())
	if err != nil {
		log.Fatal(err)
	}

	translateSrv = NewTranslateServiceServer(repository, &mockTranslator{})
	os.Exit(m.Run())
}

func Test_PopulateTranslatedMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	originalMessages1 := randOriginalMessages(5)
	originalMessages2 := randOriginalMessages(5)
	originalMessages3 := randOriginalMessages(5)

	tests := []struct {
		name               string
		originalMessages   *model.Messages
		translatedMessages []model.Messages
		assertFunc         func(t *testing.T, originalMessages *model.Messages, translatedMessages []model.Messages)
	}{
		{
			// Original messages with 5 messages, and one translated messages with same 5 messages.messages ID's.
			name:               "Populate one translation file",
			originalMessages:   originalMessages1,
			translatedMessages: randTranslatedMessages(1, 5, originalMessages1),
		},
		{
			// Original messages with 5 messages, and three translated messages with same 5 messages.messages ID's.
			name:               "Populate multiple translated messages",
			originalMessages:   originalMessages2,
			translatedMessages: randTranslatedMessages(3, 5, originalMessages2),
		},
		{
			// Original messages with 5 messages, and one empty translated messages.
			name:               "Populate empty translated messages",
			originalMessages:   originalMessages3,
			translatedMessages: randTranslatedMessages(1, 0, originalMessages3),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Add some extra messages to the original messages
			tt.originalMessages.Messages = append(tt.originalMessages.Messages, *rand.ModelMessage())
			tt.originalMessages.Messages = append(tt.originalMessages.Messages, *rand.ModelMessage())

			// Insert the original and translated messages into the repository
			service := prepareMessages(t, tt.originalMessages, tt.translatedMessages)

			// Create a slice with all messages (Original + translated)
			allMessages := append(tt.translatedMessages, *tt.originalMessages)

			// Invoke populateTranslatedMessages
			err := translateSrv.populateTranslatedMessages(ctx, service.ID, tt.originalMessages, allMessages)
			require.NoError(t, err)

			// Load updated translated messages
			loadedMsgs, err := translateSrv.repo.LoadMessages(ctx, service.ID, common.LoadMessagesOpts{})
			require.NoError(t, err, "load updated translated messages")

			// Assert that length of loaded messages is equal to the length of all messages. (one for original + count of translated messages)
			require.Len(t, loadedMsgs, len(allMessages))
			// Assert that the length of the messages in the loaded messages is equal to the length of the original messages.
			for _, m := range loadedMsgs {
				require.Len(t, m.Messages, len(tt.originalMessages.Messages))
			}
		})

	}
}

// helpers

// prepareMessages creates a service, and inserts it together with the original and translated messages into the repository.
func prepareMessages(t *testing.T, originalMessages *model.Messages, translatedMessages []model.Messages) (service *model.Service) {
	ctx := context.Background()
	service = rand.ModelService()

	err := translateSrv.repo.SaveService(ctx, service)
	require.NoError(t, err, "create test service")

	err = translateSrv.repo.SaveMessages(ctx, service.ID, originalMessages)
	require.NoError(t, err, "create original test messages")

	for _, m := range translatedMessages {
		err = translateSrv.repo.SaveMessages(ctx, service.ID, &m)
		require.NoError(t, err, "create translated test messages")
	}

	return service
}

// randOriginalMessages creates a random messages with the original flag set to true.
func randOriginalMessages(messageCount uint) *model.Messages {
	return rand.ModelMessages(messageCount, nil, rand.WithOriginal(true), rand.WithLanguage(language.English))
}

// randTranslatedMessages creates a random messages with the original flag set to false.
func randTranslatedMessages(n uint, msgCount uint, original *model.Messages) []model.Messages {
	messages := make([]model.Messages, n)
	for i, lang := range rand.Languages(n) {
		messages[i] = *rand.ModelMessages(
			msgCount,
			nil,
			rand.WithOriginal(false),
			rand.WithSameIDs(original),
			rand.WithLanguage(lang))
	}

	return messages
}
