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

func Test_SaveTranslation(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)

		tests := []struct {
			translation *model.Translation
			expectedErr error
			name        string
			serviceID   uuid.UUID
		}{
			{
				name:        "Happy path",
				serviceID:   service.ID,
				translation: rand.ModelTranslation(3, nil),
				expectedErr: nil,
			},
			{
				name:        "Missing service",
				serviceID:   uuid.New(),
				translation: rand.ModelTranslation(3, nil),
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

				require.NoError(t, err, "Save Translation")

				// Assure that the translations were saved correctly.

				actualTranslations, err := repository.LoadTranslations(ctx, tt.serviceID,
					repo.LoadTranslationOpts{FilterLanguages: []language.Tag{tt.translation.Language}})
				require.NoError(t, err, "Load saved translations")

				testutil.EqualTranslations(t, tt.translation, &actualTranslations[0])
			})
		}
	})
}

func Test_SaveTranslationsMultipleLangOneService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subTest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)

		// Create unique languages
		languages := rand.Languages(3)
		translations := make([]*model.Translation, len(languages))

		// Create translation for each language
		for i, lang := range languages {
			translations[i] = rand.ModelTranslation(3, nil, rand.WithLanguage(lang))
		}

		// Save translation
		for _, translation := range translations {
			err := repository.SaveTranslation(testCtx, service.ID, translation)
			require.NoError(t, err, "Save translation")
		}

		// Assure that all translation are saved
		for _, translation := range translations {
			actualTranslations, err := repository.LoadTranslations(testCtx, service.ID,
				repo.LoadTranslationOpts{FilterLanguages: []language.Tag{translation.Language}})
			require.NoError(t, err, "Load saved translations")

			testutil.EqualTranslations(t, translation, &actualTranslations[0])
		}
	})
}

func Test_SaveTranslationUpdate(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := prepareService(testCtx, t, repository)
		expectedTranslations := rand.ModelTranslation(3, nil)

		err := repository.SaveTranslation(testCtx, service.ID, expectedTranslations)
		require.NoError(t, err, "Save translation")

		// Update Message, Description and Status values, while keeping the ID
		for i := range expectedTranslations.Messages {
			expectedTranslations.Messages[i].Message = gofakeit.SentenceSimple()
			expectedTranslations.Messages[i].Description = gofakeit.SentenceSimple()
			expectedTranslations.Messages[i].Status = rand.MessageStatus()
		}

		// Save updated translations

		err = repository.SaveTranslation(testCtx, service.ID, expectedTranslations)
		require.NoError(t, err, "Update messages")

		// Assure that translations are updated

		actualTranslation, err := repository.LoadTranslations(testCtx, service.ID,
			repo.LoadTranslationOpts{FilterLanguages: []language.Tag{expectedTranslations.Language}})
		require.NoError(t, err, "Load updated translations")

		testutil.EqualTranslations(t, expectedTranslations, &actualTranslation[0])
	})
}

func Test_LoadTranslation(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		langs := rand.Languages(2)
		// translationLang is the language of the translation, for Happy Path test.
		translationLang := langs[0]
		// langWithNoTranslations is a different language, for No translation with language test.
		// It must be different from translationLang, since otherwise that would return not nil Translation.Messages
		// and will fail the test.
		langWithNoTranslations := langs[1]

		// Prepare
		service := prepareService(testCtx, t, repository)
		translations := rand.ModelTranslation(3, nil, rand.WithLanguage(translationLang))

		err := repository.SaveTranslation(testCtx, service.ID, translations)
		require.NoError(t, err, "Prepare test translation")

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
				name:      "No translations with service",
				expected:  []model.Translation{},
				serviceID: uuid.New(),
			},
			{
				language:  langWithNoTranslations,
				name:      "No translations with language",
				expected:  []model.Translation{},
				serviceID: service.ID,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actualTranslations, err := repository.LoadTranslations(ctx, tt.serviceID,
					repo.LoadTranslationOpts{FilterLanguages: []language.Tag{tt.language}})
				require.NoError(t, err, "Load translations")

				assert.ElementsMatch(t, tt.expected, actualTranslations)
			})
		}
	})
}

func Test_LoadAllTranslationsForService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare

		service := prepareService(testCtx, t, repository)
		languages := rand.Languages(gofakeit.UintRange(1, 5))
		translations := make([]model.Translation, 0, len(languages))

		for _, lang := range languages {
			translation := rand.ModelTranslation(1, nil, rand.WithLanguage(lang))
			err := repository.SaveTranslation(testCtx, service.ID, translation)
			require.NoError(t, err, "Prepare test translations")
			translations = append(translations, *translation)
		}

		tests := []struct {
			name                 string
			expectedTranslations []model.Translation
			languages            []language.Tag
			serviceID            uuid.UUID
		}{
			{
				name:                 "Happy Path, all service translations",
				expectedTranslations: translations,
				serviceID:            service.ID,
			},
			{
				name:                 "Happy Path, filter by existing languages",
				expectedTranslations: translations,
				serviceID:            service.ID,
				languages:            languages,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actualTranslations, err := repository.LoadTranslations(ctx, tt.serviceID,
					repo.LoadTranslationOpts{FilterLanguages: tt.languages})

				require.NoError(t, err, "Load translations")
				assert.ElementsMatch(t, actualTranslations, tt.expectedTranslations)
			})
		}
	})
}
