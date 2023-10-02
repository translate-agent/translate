//go:build integration

package factory

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
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
			translation    *model.Translation
			expectedErr error
			name        string
			serviceID   uuid.UUID
		}{
			{
				name:        "Happy path",
				serviceID:   service.ID,
				translation:    rand.ModelTranslation(3, nil),
				expectedErr: nil,
			},
			{
				name:        "Missing service",
				serviceID:   uuid.New(),
				translation:    rand.ModelTranslation(3, nil),
				expectedErr: repo.ErrNotFound,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				err := repository.SaveTranslation(ctx, tt.serviceID, tt.translation)

				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
					return
				}

				require.NoError(t, err, "Save messages")

				// Assure that the messages were saved correctly.

				actualMessages, err := repository.LoadTranslation(ctx, tt.serviceID,
					repo.LoadTranslationOpts{FilterLanguages: []language.Tag{tt.translation.Language}})
				require.NoError(t, err, "Load saved messages")

				testutil.EqualTranslations(t, tt.translation, &actualMessages[0])
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
		languages := rand.Languages(3)
		translations := make([]*model.Translation, len(languages))

		// Create messages for each language
		for i, lang := range languages {
			translations[i] = rand.ModelTranslation(3, nil, rand.WithLanguage(lang))
		}

		// Save messages
		for _, m := range translations {
			err := repository.SaveTranslation(testCtx, service.ID, m)
			require.NoError(t, err, "Save messages")
		}

		// Assure that all messages are saved
		for _, m := range translations {
			actualMessages, err := repository.LoadTranslation(testCtx, service.ID,
				repo.LoadTranslationOpts{FilterLanguages: []language.Tag{m.Language}})
			require.NoError(t, err, "Load saved messages")

			testutil.EqualTranslations(t, m, &actualMessages[0])
		}
	})
}

func Test_SaveMessagesUpdate(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)
		expectedMessages := rand.ModelTranslation(3, nil)

		err := repository.SaveTranslation(testCtx, service.ID, expectedMessages)
		require.NoError(t, err, "Save messages")

		// Update Message, Description and Status values, while keeping the ID
		for i := range expectedMessages.Messages {
			expectedMessages.Messages[i].Message = gofakeit.SentenceSimple()
			expectedMessages.Messages[i].Description = gofakeit.SentenceSimple()
			expectedMessages.Messages[i].Status = rand.MessageStatus()
		}

		// Save updated messages

		err = repository.SaveTranslation(testCtx, service.ID, expectedMessages)
		require.NoError(t, err, "Update messages")

		// Assure that messages are updated

		actualMessages, err := repository.LoadTranslation(testCtx, service.ID,
			repo.LoadTranslationOpts{FilterLanguages: []language.Tag{expectedMessages.Language}})
		require.NoError(t, err, "Load updated messages")

		testutil.EqualTranslations(t, expectedMessages, &actualMessages[0])
	})
}

func Test_LoadMessages(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		langs := rand.Languages(2)
		// messagesLang is the language of the messages, for Happy Path test.
		messagesLang := langs[0]
		// langWithNoMsgs is a different language, for No messages with language test.
		// It must be different from messagesLang, since otherwise that would return not nil Messages.Messages
		// and will fail the test.
		langWithNoMsgs := langs[1]

		// Prepare
		service := prepareService(testCtx, t, repository)
		translations := rand.ModelTranslation(3, nil, rand.WithLanguage(messagesLang))

		err := repository.SaveTranslation(testCtx, service.ID, translations)
		require.NoError(t, err, "Prepare test messages")

		tests := []struct {
			language  language.Tag
			name      string
			expected  []model.Translation
			serviceID uuid.UUID
		}{
			{
				language:  translations.Language,
				name:      "Happy Path",
				expected:  []model.Translation{*translations},
				serviceID: service.ID,
			},
			{
				language:  translations.Language,
				name:      "No messages with service",
				expected:  []model.Translation{},
				serviceID: uuid.New(),
			},
			{
				language:  langWithNoMsgs,
				name:      "No messages with language",
				expected:  []model.Translation{},
				serviceID: service.ID,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actualMessages, err := repository.LoadTranslation(ctx, tt.serviceID,
					repo.LoadTranslationOpts{FilterLanguages: []language.Tag{tt.language}})
				require.NoError(t, err, "Load messages")

				assert.ElementsMatch(t, tt.expected, actualMessages)
			})
		}
	})
}

func Test_LoadAllMessagesForService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare

		service := prepareService(testCtx, t, repository)
		languages := rand.Languages(gofakeit.UintRange(1, 5))
		translations := make([]model.Translation, 0, len(languages))

		for _, lang := range languages {
			msgs := rand.ModelTranslation(1, nil, rand.WithLanguage(lang))
			err := repository.SaveTranslation(testCtx, service.ID, msgs)
			require.NoError(t, err, "Prepare test messages")
			translations = append(translations, *msgs)
		}

		tests := []struct {
			name         string
			expectedMsgs []model.Translation
			languages    []language.Tag
			serviceID    uuid.UUID
		}{
			{
				name:         "Happy Path, all service messages",
				expectedMsgs: translations,
				serviceID:    service.ID,
			},
			{
				name:         "Happy Path, filter by existing languages",
				expectedMsgs: translations,
				serviceID:    service.ID,
				languages:    languages,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actualMessages, err := repository.LoadTranslation(ctx, tt.serviceID,
					repo.LoadTranslationOpts{FilterLanguages: tt.languages})

				require.NoError(t, err, "Load messages")
				assert.ElementsMatch(t, actualMessages, tt.expectedMsgs)
			})
		}
	})
}
