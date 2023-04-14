//go:build integration

package mysql

import (
	"context"
	"reflect"
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

func prepareService(ctx context.Context, t *testing.T) *model.Service {
	t.Helper()

	service := randService()

	err := repository.SaveService(ctx, service)
	require.NoError(t, err, "Prepare test service")

	return service
}

func assertEqualTranslationFile(t *testing.T, expected, actual *model.TranslationFile) {
	t.Helper()

	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Messages.Language, actual.Messages.Language)
	assert.ElementsMatch(t, expected.Messages.Messages, actual.Messages.Messages)
}

func Test_SaveTranslationFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := prepareService(ctx, t)

	// Translation files

	happyTranslationFileWithUUID := randTranslationFile(randMessages())

	happyTranslationFileWithoutUUID := randTranslationFile(randMessages())
	happyTranslationFileWithoutUUID.ID = uuid.Nil

	missingServiceTranslationFile := randTranslationFile(randMessages())

	tests := []struct {
		name            string
		serviceID       uuid.UUID
		translationFile *model.TranslationFile
		expectedErr     error
	}{
		{
			name:            "Happy path with UUID",
			serviceID:       service.ID,
			translationFile: happyTranslationFileWithUUID,
			expectedErr:     nil,
		},
		{
			name:            "Happy path without UUID",
			serviceID:       service.ID,
			translationFile: happyTranslationFileWithoutUUID,
			expectedErr:     nil,
		},
		{
			name:            "Missing service",
			serviceID:       uuid.New(),
			translationFile: missingServiceTranslationFile,
			expectedErr:     &repo.NotFoundError{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repository.SaveTranslationFile(ctx, tt.serviceID, tt.translationFile)

			if tt.expectedErr != nil {
				e := reflect.New(reflect.TypeOf(tt.expectedErr).Elem()).Interface()
				assert.ErrorAs(t, err, &e)

				return
			}

			require.NoError(t, err, "Save translation file")

			// Check if is inserted

			actualTranslationFile, err := repository.LoadTranslationFile(ctx, tt.serviceID, tt.translationFile.Messages.Language)
			require.NoError(t, err, "Load saved translation file")

			assertEqualTranslationFile(t, tt.translationFile, actualTranslationFile)
		})

	}
}

func Test_UpdateTranslationFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := prepareService(ctx, t)

	expectedTranslationFile := randTranslationFile(randMessages())

	err := repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
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

func Test_LoadTranslationFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := prepareService(ctx, t)

	expectedTranslationFile := randTranslationFile(randMessages())

	err := repository.SaveTranslationFile(ctx, service.ID, expectedTranslationFile)
	require.NoError(t, err, "Save translation file")

	tests := []struct {
		translationFile *model.TranslationFile
		expectedErr     error
		name            string
		serviceID       uuid.UUID
	}{
		{
			name:            "All OK",
			translationFile: expectedTranslationFile,
			serviceID:       service.ID,
			expectedErr:     nil,
		},
		{
			name:        "Nonexistent",
			serviceID:   uuid.New(),
			expectedErr: &repo.NotFoundError{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualTranslationFile, err := repository.LoadTranslationFile(
				ctx,
				tt.serviceID,
				expectedTranslationFile.Messages.Language,
			)

			if tt.expectedErr != nil {
				e := reflect.New(reflect.TypeOf(tt.expectedErr).Elem()).Interface()
				assert.ErrorAs(t, err, &e)

				return
			}

			require.NoError(t, err)

			assertEqualTranslationFile(t, tt.translationFile, actualTranslationFile)
		})
	}
}
