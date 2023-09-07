package server

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"strings"
	"testing"

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

	originalMessages1 := randOriginalMessages(5)
	translatedMessages := randTranslatedMessages(5, 5, originalMessages1)
	mixedMessages := append(translatedMessages, *originalMessages1)

	tests := []struct {
		name            string
		messages        model.MessagesSlice
		untranslatedIds []string
		expected        model.MessagesSlice
	}{
		{
			name:            "No untranslated IDs are provided",
			messages:        model.MessagesSlice{*originalMessages1},
			untranslatedIds: nil,
			expected:        model.MessagesSlice{*originalMessages1},
		},
		{
			name:            "Single original message, untranslated IDs are provided",
			messages:        model.MessagesSlice{*originalMessages1},
			untranslatedIds: []string{originalMessages1.Messages[0].ID},
			expected:        model.MessagesSlice{*originalMessages1},
		},
		{
			name:            "Non originals messages, untranslated IDs are provided",
			messages:        translatedMessages,
			untranslatedIds: []string{translatedMessages[0].Messages[0].ID},
			expected:        translatedMessages,
		},
		{
			name:            "Update altered messages for transaction ",
			messages:        mixedMessages,
			untranslatedIds: []string{originalMessages1.Messages[0].ID},
			expected:        mixedMessages,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := translateSrv.alterTranslations(tt.messages, tt.untranslatedIds)
			require.Len(t, result, len(tt.expected))

			original, _ := tt.messages.SplitOriginal()

			if original == nil || len(tt.expected) == 1 || len(tt.untranslatedIds) == 0 {
				require.Equal(t, tt.expected, tt.messages)
				return
			}

			// Check that messages have been altered and marked as untranslated for all translations.
			for i, m := range result {
				if m.Original {
					require.Equal(t, tt.messages[i], m)
					continue
				}

				for _, msg := range m.Messages {
					if slices.Contains(tt.untranslatedIds, msg.ID) {
						require.Equal(t, msg.Status, model.MessageStatusUntranslated)
					}
				}

			}
		})
	}
}

func Test_populateTranslations(t *testing.T) {
	t.Parallel()

	originalMessages1 := randOriginalMessages(5)
	originalMessages2 := randOriginalMessages(5)
	originalMessages3 := randOriginalMessages(5)

	tests := []struct {
		name               string
		originalMessages   *model.Messages
		translatedMessages []model.Messages
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

			// Invoke populateTranslations
			newMessages := translateSrv.populateTranslations(allMessages)

			// Assert that length of loaded messages is equal to the length of all messages. (one for original + count of translated messages)
			require.Len(t, newMessages, len(allMessages))

			// Assert that the length of the messages in the loaded messages is equal to the length of the original messages.
			for i, m := range newMessages {
				if m.Original {
					require.Equal(t, *tt.originalMessages, m)
					continue
				}

				require.Len(t, m.Messages, len(newMessages[i].Messages))
			}
		})
	}
}

func Test_fuzzyTranslate(t *testing.T) {
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
			newMessages, err := translateSrv.fuzzyTranslate(ctx, allMessages)
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
