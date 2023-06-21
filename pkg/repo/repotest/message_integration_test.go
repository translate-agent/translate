//go:build integration

package repotest

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/repo/common"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

func prepareService(ctx context.Context, t *testing.T, repository repo.Repo) *model.Service {
	t.Helper()

	service := rand.ModelService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	return service
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
				messages:    rand.ModelMessages(3, nil),
				expectedErr: nil,
			},
			{
				name:        "Missing service",
				serviceID:   uuid.New(),
				messages:    rand.ModelMessages(3, nil),
				expectedErr: common.ErrNotFound,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				err := repository.SaveMessages(ctx, tt.serviceID, tt.messages)

				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
					return
				}

				require.NoError(t, err, "Save messages")

				// Assure that the messages were saved correctly.

				actualMessages, err := repository.LoadMessages(ctx, tt.serviceID,
					common.LoadMessagesOpts{FilterLanguages: []language.Tag{tt.messages.Language}})
				require.NoError(t, err, "Load saved messages")

				testutil.EqualMessages(t, tt.messages, &actualMessages[0])
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

		// Create unique languages
		langs := rand.Langs(3)
		messages := make([]*model.Messages, len(langs))

		// Create messages for each language
		for i, lang := range langs {
			messages[i] = rand.ModelMessages(3, nil, rand.WithLanguage(lang))
		}

		// Save messages
		for _, m := range messages {
			err := repository.SaveMessages(testCtx, service.ID, m)
			require.NoError(t, err, "Save messages")
		}

		// Assure that all messages are saved
		for _, m := range messages {
			actualMessages, err := repository.LoadMessages(testCtx, service.ID,
				common.LoadMessagesOpts{FilterLanguages: []language.Tag{m.Language}})
			require.NoError(t, err, "Load saved messages")

			testutil.EqualMessages(t, m, &actualMessages[0])
		}
	})
}

func Test_SaveMessagesUpdate(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)
		expectedMessages := rand.ModelMessages(3, nil)

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

		actualMessages, err := repository.LoadMessages(testCtx, service.ID,
			common.LoadMessagesOpts{FilterLanguages: []language.Tag{expectedMessages.Language}})
		require.NoError(t, err, "Load updated messages")

		testutil.EqualMessages(t, expectedMessages, &actualMessages[0])
	})
}

func Test_LoadMessages(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		langs := rand.Langs(2)
		// messagesLang is the language of the messages, for Happy Path test.
		messagesLang := langs[0]
		// langWithNoMsgs is a different language, for No messages with language test.
		// It must be different from messagesLang, since otherwise that would return not nil Messages.Messages
		// and will fail the test.
		langWithNoMsgs := langs[1]

		// Prepare
		service := prepareService(testCtx, t, repository)
		messages := rand.ModelMessages(3, nil, rand.WithLanguage(messagesLang))

		err := repository.SaveMessages(testCtx, service.ID, messages)
		require.NoError(t, err, "Prepare test messages")

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
				serviceID: uuid.New(),
				language:  messages.Language,
				expected:  &model.Messages{Language: messagesLang},
			},
			{
				name:      "No messages with language",
				serviceID: service.ID,
				language:  langWithNoMsgs,
				expected:  &model.Messages{Language: langWithNoMsgs},
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actualMessages, err := repository.LoadMessages(ctx, tt.serviceID,
					common.LoadMessagesOpts{FilterLanguages: []language.Tag{tt.language}})
				require.NoError(t, err, "Load messages")

				testutil.EqualMessages(t, tt.expected, &actualMessages[0])
			})
		}
	})
}
