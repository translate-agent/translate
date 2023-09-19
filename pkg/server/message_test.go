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
	"golang.org/x/exp/slices"
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

func Test_populateTranslations(t *testing.T) {
	t.Parallel()

	originalMessages1 := randOriginalMessages(5)
	translatedMessages1 := randTranslatedMessages(1, uint(len(originalMessages1.Messages)), originalMessages1)

	originalMessages2 := randOriginalMessages(5)
	translatedMessages2 := randTranslatedMessages(2, uint(len(originalMessages2.Messages)), originalMessages2)
	originalMessages2.Messages = append(originalMessages2.Messages, *rand.ModelMessage())

	originalMessages3 := randOriginalMessages(5)

	tests := []struct {
		name               string
		originalMessages   *model.Messages
		translatedMessages []model.Messages
	}{
		{
			// Original messages and translated messages, have same messages.messages ID's -> noop
			name:               "Nothing to populate",
			originalMessages:   originalMessages1,
			translatedMessages: translatedMessages1,
		},
		{
			// Original messages have one extra message -> translated messages should be populated with the extra message.
			name:               "Populate two translations with one new message",
			originalMessages:   originalMessages2,
			translatedMessages: translatedMessages2,
		},
		{
			// Translated messages are empty -> translated messages should be populated with all messages from original messages.
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
			allMessages := model.MessagesSlice{*tt.originalMessages}
			allMessages = append(allMessages, tt.translatedMessages...)

			newMessages := translateSrv.populateTranslations(allMessages)

			// Assert that length of loaded messages is equal to the length of all messages.
			require.Len(t, newMessages, len(allMessages))

			original, others := newMessages.SplitOriginal()

			// Assert that original messages are not altered.
			require.Equal(t, tt.originalMessages, original)

			// Assert that all messages from original messages are present in translated messages.
			for _, m := range others {
				require.Len(t, m.Messages, len(tt.originalMessages.Messages))

				// Check each message individually, for correct population.
				for _, origMsg := range tt.originalMessages.Messages {
					require.True(t, slices.ContainsFunc(m.Messages, func(transMsg model.Message) bool {
						return transMsg.ID == origMsg.ID
					}), "translated messages should contain original message with ID %s", origMsg.ID)
				}
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
			untranslatedMessageIDLookup := make(map[string]struct{})

			// Randomly set message status to untranslated
			for _, msg := range tt.originalMessages.Messages {
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

func Test_getUntranslatedIDs(t *testing.T) {
	old := &model.Messages{
		Messages: []model.Message{
			{ID: "1", Message: "Hello"},
			{ID: "2", Message: "World"},
		},
	}
	new := &model.Messages{
		Messages: []model.Message{
			{ID: "1", Message: "Hello"},
			{ID: "2", Message: "Go"},
			{ID: "3", Message: "Testing"},
		},
	}

	changedIDs := getUntranslatedIDs(old, new)

	// ID:1 -> Are the same (Should not be included)
	// ID:2 -> Messages has been changed (Should be included)
	// ID:3 -> Is new (Should be included)
	require.Equal(t, []string{"2", "3"}, changedIDs)
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
	return rand.ModelMessages(
		messageCount,
		[]rand.ModelMessageOption{rand.WithStatus(model.MessageStatusTranslated)},
		rand.WithOriginal(true),
		rand.WithLanguage(language.English))
}

// randTranslatedMessages creates a random messages with the original flag set to false
// with the same IDs as the original messages, and with translated status.
func randTranslatedMessages(n uint, msgCount uint, original *model.Messages) []model.Messages {
	messages := make([]model.Messages, n)
	for i, lang := range rand.Languages(n) {
		messages[i] = *rand.ModelMessages(
			msgCount,
			[]rand.ModelMessageOption{rand.WithStatus(model.MessageStatusTranslated)},
			rand.WithOriginal(false),
			rand.WithSameIDs(original),
			rand.WithLanguage(lang))
	}

	return messages
}
