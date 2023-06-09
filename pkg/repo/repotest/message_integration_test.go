//go:build integration

package repotest

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/testutil"
	"golang.org/x/text/language"
)

func randMessages() *model.Messages {
	size := gofakeit.IntRange(0, 5)

	messages := make([]model.Message, 0, size)

	for i := 0; i < size; i++ {
		messages = append(messages, model.Message{
			ID:          gofakeit.SentenceSimple(),
			Message:     gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
			Fuzzy:       gofakeit.Bool(),
		},
		)
	}

	return &model.Messages{Language: language.MustParse(gofakeit.LanguageBCP()), Messages: messages}
}

func prepareService(ctx context.Context, t *testing.T, repository repo.Repo) *model.Service {
	t.Helper()

	service := randService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	return service
}

func requireEqualMessages(t *testing.T, expected, actual *model.Messages) {
	t.Helper()

	require.Equal(t, expected.Language, actual.Language)
	require.ElementsMatch(t, expected.Messages, actual.Messages)
}

func Test_SaveMessages(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)

		tests := []struct {
			messages    *model.Messages
			expectedErr error
			name        string
			serviceID   uuid.UUID
		}{
			{
				name:        "Happy path",
				serviceID:   service.ID,
				messages:    randMessages(),
				expectedErr: nil,
			},
			{
				name:        "Missing service",
				serviceID:   uuid.New(),
				messages:    randMessages(),
				expectedErr: repo.ErrNotFound,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				fmt.Printf("tt.serviceID: %v\n", tt.serviceID)
				err := repository.SaveMessages(ctx, tt.serviceID, tt.messages)

				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
					return
				}

				require.NoError(t, err, "Save messages")

				// Assure that the messages were saved correctly.

				actualMessages, err := repository.LoadMessages(ctx, tt.serviceID, tt.messages.Language)
				require.NoError(t, err, "Load saved messages")

				requireEqualMessages(t, tt.messages, actualMessages)
			})
		}
	})
}

func Test_SaveMessagesMultipleLangOneService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subTest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)

		count := gofakeit.IntRange(3, 5)
		messages := make([]*model.Messages, 0, count)

		languagesUsed := make(map[language.Tag]bool, count)

		// Create messages with different languages
		for i := 0; i < count; i++ {
			msg := randMessages()

			// Make sure we don't use the same language twice.
			for languagesUsed[msg.Language] {
				msg = randMessages()
			}

			languagesUsed[msg.Language] = true

			messages = append(messages, msg)
		}

		// Save messages
		for _, m := range messages {
			err := repository.SaveMessages(testCtx, service.ID, m)
			require.NoError(t, err, "Save messages")
		}

		// Assure that all messages are saved
		for _, m := range messages {
			actualMessages, err := repository.LoadMessages(testCtx, service.ID, m.Language)
			require.NoError(t, err, "Load saved messages")

			requireEqualMessages(t, m, actualMessages)
		}
	})
}

func Test_SaveMessagesUpdate(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)
		expectedMessages := randMessages()

		err := repository.SaveMessages(testCtx, service.ID, expectedMessages)
		require.NoError(t, err, "Save messages")

		// Update Message, Description and Fuzzy values, while keeping the ID
		for i := range expectedMessages.Messages {
			expectedMessages.Messages[i].Message = gofakeit.SentenceSimple()
			expectedMessages.Messages[i].Description = gofakeit.SentenceSimple()
			expectedMessages.Messages[i].Fuzzy = gofakeit.Bool()
		}

		// Save updated messages

		err = repository.SaveMessages(testCtx, service.ID, expectedMessages)
		require.NoError(t, err, "Update messages")

		// Assure that messages are updated

		actualMessages, err := repository.LoadMessages(testCtx, service.ID, expectedMessages.Language)
		require.NoError(t, err, "Load updated messages")

		requireEqualMessages(t, expectedMessages, actualMessages)
	})
}

func Test_LoadMessages(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)
		messages := randMessages()

		err := repository.SaveMessages(testCtx, service.ID, messages)
		require.NoError(t, err, "Prepare test messages")

		missingServiceID := uuid.New()
		missingLang := language.MustParse(gofakeit.LanguageBCP())
		// Make sure we don't use the same language as the message, since that would return not nil Messages.Messages.
		for missingLang == messages.Language {
			missingLang = language.MustParse(gofakeit.LanguageBCP())
		}

		tests := []struct {
			expected  *model.Messages
			language  language.Tag
			name      string
			serviceID uuid.UUID
		}{
			{
				name:      "Happy Path",
				expected:  messages,
				serviceID: service.ID,
				language:  messages.Language,
			},
			{
				name:      "No messages with service",
				serviceID: missingServiceID,
				language:  messages.Language,
				expected:  &model.Messages{Language: messages.Language, Messages: nil},
			},
			{
				name:      "No messages with language",
				serviceID: service.ID,
				language:  missingLang,
				expected:  &model.Messages{Language: missingLang, Messages: nil},
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actualMessages, err := repository.LoadMessages(ctx, tt.serviceID, tt.language)
				require.NoError(t, err, "Load messages")

				requireEqualMessages(t, tt.expected, actualMessages)
			})
		}
	})
}