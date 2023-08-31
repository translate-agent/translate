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

func Test_alterTranslations(t *testing.T) {
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

			allMessages := append(tt.translatedMessages, *tt.originalMessages)

			// Prepare new original messages with randomly altered message text

			newOriginalMessages := &model.MessagesSlice{*tt.originalMessages}.Clone()[0]
			alteredMessageLookup := make(map[string]string)

			for i, msg := range newOriginalMessages.Messages {
				if gofakeit.Bool() {
					newOriginalMessages.Messages[i].Message = gofakeit.BuzzWord()
					alteredMessageLookup[msg.ID] = newOriginalMessages.Messages[i].Message
				}
			}

			// Invoke alterTranslations
			newMessages, err := translateSrv.alterTranslations(ctx, allMessages, newOriginalMessages)
			require.NoError(t, err)
			require.Len(t, newMessages, len(allMessages))

			// Check that messages have been altered and marked as untranslated for all translations.
			for _, m := range newMessages {
				if m.Original {
					require.Equal(t, *tt.originalMessages, m)
					continue
				}

				for _, message := range m.Messages {
					if alteredMessage, ok := alteredMessageLookup[message.ID]; ok {
						require.Equal(t, alteredMessage, message.Message)
						require.Equal(t, model.MessageStatusUntranslated, message.Status)
					}
				}
			}
		})
	}
}

func Test_populateTranslations(t *testing.T) {
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

			// Create a slice with all messages (Original + translated)
			allMessages := append(tt.translatedMessages, *tt.originalMessages)

			// Add some extra messages to the new original messages
			newOriginalMessages := &model.MessagesSlice{*tt.originalMessages}.Clone()[0]
			newOriginalMessages.Messages = append(newOriginalMessages.Messages, *rand.ModelMessage())
			newOriginalMessages.Messages = append(newOriginalMessages.Messages, *rand.ModelMessage())

			// Invoke populateTranslations
			newMessages, err := translateSrv.populateTranslations(ctx, allMessages, newOriginalMessages)
			require.NoError(t, err)

			// Assert that length of loaded messages is equal to the length of all messages. (one for original + count of translated messages)
			require.Len(t, newMessages, len(allMessages))

			// Assert that the length of the messages in the loaded messages is equal to the length of the original messages.
			for _, m := range newMessages {
				if m.Original {
					require.Equal(t, *tt.originalMessages, m)
					continue
				}

				require.Len(t, m.Messages, len(newOriginalMessages.Messages))
			}
		})
	}
}

func Test_refreshTranslations(t *testing.T) {
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
			name:               "Fuzzy translate untranslated messages for one translation",
			originalMessages:   originalMessages1,
			translatedMessages: randTranslatedMessages(1, 3, originalMessages1),
		},
		{
			name:               "Fuzzy translate untranslated messages for five translations",
			originalMessages:   originalMessages2,
			translatedMessages: randTranslatedMessages(5, 5, originalMessages2),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			allMessages := append(tt.translatedMessages, *tt.originalMessages)
			newOriginalMessages := &model.MessagesSlice{*tt.originalMessages}.Clone()[0]
			untranslatedMessageIDLookup := make(map[string]struct{})

			// Randomly set message status to untranslated

			for _, msg := range newOriginalMessages.Messages {
				if gofakeit.Bool() {
					untranslatedMessageIDLookup[msg.ID] = struct{}{}
				}
			}

			for i := range allMessages {
				if allMessages[i].Original {
					continue
				}

				for j := range allMessages[i].Messages {
					if _, ok := untranslatedMessageIDLookup[allMessages[i].Messages[j].ID]; ok {
						allMessages[i].Messages[j].Status = model.MessageStatusUntranslated
					}
				}
			}

			// Invoke refreshTranslations
			newMessages, err := translateSrv.refreshTranslations(ctx, allMessages, newOriginalMessages)
			require.NoError(t, err)

			require.Len(t, newMessages, len(allMessages))

			// Check that untranslated messages have been translated and marked as fuzzy for all translations.
			for _, m := range newMessages {
				if m.Original {
					require.Equal(t, *tt.originalMessages, m)
					continue
				}

				for _, message := range m.Messages {
					if _, ok := untranslatedMessageIDLookup[message.ID]; ok {
						require.Equal(t, mockTranslation, message.Message)
						require.Equal(t, model.MessageStatusFuzzy, message.Status)
					}
				}
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
