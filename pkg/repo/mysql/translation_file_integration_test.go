//go:build integration

package mysql

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
)

func randMessages() []model.Message {
	size := gofakeit.IntRange(0, 5)

	messages := make([]model.Message, 0, size)

	for i := 0; i < size; i++ {
		messages = append(messages, model.Message{
			ID:          gofakeit.Word(),
			Message:     gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
			Fuzzy:       gofakeit.Bool(),
		},
		)
	}

	return messages
}

func randTranslationFile(messages []model.Message) *model.TranslationFile {
	return &model.TranslationFile{
		ID:       uuid.New(),
		Messages: model.Messages{Messages: messages, Language: language.MustParse(gofakeit.LanguageBCP())},
	}
}

func assertEqualTranslationFile(t *testing.T, expected, actual *model.TranslationFile) {
	t.Helper()

	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Messages.Language, actual.Messages.Language)
	require.ElementsMatch(t, expected.Messages.Messages, actual.Messages.Messages)
}

func Test_SaveTranslationFileWithUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	// Actual test

	expectedTranslationFile := randTranslationFile(randMessages())

	err = repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
	require.NoError(t, err, "Save translation file")

	// Check if is inserted

	actualTranslationFile, err := repository.LoadTranslationFile(
		ctx,
		service.ID,
		expectedTranslationFile.Messages.Language,
	)
	require.NoError(t, err, "Load translation file")

	assertEqualTranslationFile(t, expectedTranslationFile, actualTranslationFile)
}

func Test_SaveTranslationFileWithoutUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	// Actual Test

	expectedTranslationFile := randTranslationFile(randMessages())

	expectedTranslationFile.ID = uuid.Nil

	err = repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
	require.NoError(t, err, "Save translation file")

	// Check if is inserted

	actualTranslationFile, err := repository.LoadTranslationFile(
		ctx,
		service.ID,
		expectedTranslationFile.Messages.Language,
	)
	require.NoError(t, err, "Load translation file")

	assertEqualTranslationFile(t, expectedTranslationFile, actualTranslationFile)
}

func Test_UpdateTranslationFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	expectedTranslationFile := randTranslationFile(randMessages())

	err = repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
	require.NoError(t, err, "Save translation file")

	// Actual Test

	expectedTranslationFile.Messages.Messages = randMessages()

	err = repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
	require.NoError(t, err, "Update translation file")

	// Check if updated

	actualTranslationFile, err := repository.LoadTranslationFile(
		ctx,
		service.ID,
		expectedTranslationFile.Messages.Language,
	)
	require.NoError(t, err, "Load updated translation file")

	assertEqualTranslationFile(t, expectedTranslationFile, actualTranslationFile)
}

func Test_SaveTranslationFileNoService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	translationFile := randTranslationFile(randMessages())

	err := repository.SaveTranslationFile(ctx, uuid.New(), translationFile)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func Test_LoadTranslationFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	expectedTranslationFile := randTranslationFile(randMessages())

	err = repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
	require.NoError(t, err, "Save translation file")

	tests := []struct {
		expected    *model.TranslationFile
		expectedErr error
		name        string
		serviceID   uuid.UUID
	}{
		{
			name:        "All OK",
			expected:    expectedTranslationFile,
			serviceID:   service.ID,
			expectedErr: nil,
		},
		{
			name:        "Nonexistent",
			serviceID:   uuid.New(),
			expectedErr: repo.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := repository.LoadTranslationFile(
				ctx,
				tt.serviceID,
				expectedTranslationFile.Messages.Language,
			)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assertEqualTranslationFile(t, tt.expected, actual)
		})
	}
}
