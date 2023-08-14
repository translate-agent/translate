package server

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo/badgerdb"
	"go.expect.digital/translate/pkg/repo/common"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

const mockTranslation = "{Translated}"

var translateSrv *TranslateServiceServer

type mockTranslator struct{}

func (m *mockTranslator) Translate(ctx context.Context, messages *model.Messages, targetLanguage language.Tag) (*model.Messages, error) {
	newMessages := &model.Messages{
		Language: targetLanguage,
		Messages: make([]model.Message, 0, len(messages.Messages)),
		Original: messages.Original,
	}

	newMessages.Messages = append(newMessages.Messages, messages.Messages...)

	for i := range newMessages.Messages {
		newMessages.Messages[i].Message = mockTranslation
		newMessages.Messages[i].Status = model.MessageStatusFuzzy
	}

	return newMessages, nil
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

func Test_UpdateAlteredMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	originalMessages1 := randOriginalMessages(3)
	originalMessages2 := randOriginalMessages(10)

	tests := []struct {
		name               string
		originalMessages   *model.Messages
		translatedMessages []model.Messages
		assertFunc         func(t *testing.T, originalMessages *model.Messages, translatedMessages []model.Messages)
	}{
		{
			name:               "Update altered messages for one translation",
			originalMessages:   originalMessages1,
			translatedMessages: randTranslatedMessages(1, 3, originalMessages1),
		},
		{
			name:               "Update altered messages for five translations",
			originalMessages:   originalMessages2,
			translatedMessages: randTranslatedMessages(5, 5, originalMessages2),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// insert the original and translated messages into the repository
			service := prepareMessages(t, tt.originalMessages, tt.translatedMessages)
			allMessages := append(tt.translatedMessages, *tt.originalMessages)

			newOriginalMessages := &model.Messages{
				Language: tt.originalMessages.Language,
				Messages: make([]model.Message, 0, len(tt.originalMessages.Messages)),
				Original: true,
			}

			newOriginalMessages.Messages = append(newOriginalMessages.Messages, tt.originalMessages.Messages...)

			// randomly alter messages in the original language
			alteredMessageLookup := make(map[string]struct{})

			for i, msg := range newOriginalMessages.Messages {
				if gofakeit.Bool() {
					newOriginalMessages.Messages[i].Message = gofakeit.BuzzWord()
					alteredMessageLookup[msg.ID] = struct{}{}
				}
			}

			err := translateSrv.updateAlteredMessages(ctx, service.ID, allMessages, newOriginalMessages)
			require.NoError(t, err)

			// load updated translated messages
			loadedMsgs, err := translateSrv.repo.LoadMessages(ctx, service.ID, common.LoadMessagesOpts{})

			require.NoError(t, err, "load updated translated messages")
			require.ElementsMatch(t, allMessages, loadedMsgs)

			// check that altered messages have been translated and marked as fuzzy for all translations.
			for _, messages := range loadedMsgs {
				if messages.Original {
					continue
				}

				for _, message := range messages.Messages {
					if _, ok := alteredMessageLookup[message.ID]; ok {
						require.Equal(t, mockTranslation, message.Message)
						require.Equal(t, model.MessageStatusFuzzy, message.Status)
					}
				}
			}
		})
	}
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
