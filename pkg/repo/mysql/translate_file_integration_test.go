//go:build integration

package mysql

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func randTranslateFile(messages []model.Message) *model.TranslateFile {
	return &model.TranslateFile{
		ID:       uuid.New(),
		Messages: model.Messages{Messages: messages},
		Language: language.MustParse(gofakeit.LanguageBCP()),
	}
}

func assertEqualTranslateFile(t *testing.T, expected, actual *model.TranslateFile) bool {
	t.Helper()

	if eq := assert.Equal(t, expected.ID, actual.ID); !eq {
		return eq
	}

	if eq := assert.Equal(t, expected.Language, actual.Language); !eq {
		return eq
	}

	if eq := assert.ElementsMatch(t, expected.Messages.Messages, actual.Messages.Messages); !eq {
		return eq
	}

	return true
}

func Test_SaveTranslateFileWithUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "Prepare test service") {
		return
	}

	// Actual test

	expectedTranslateFile := randTranslateFile(randMessages())

	err = repository.SaveTranslateFile(ctx, service.ID, expectedTranslateFile)
	if !assert.NoError(t, err) {
		return
	}

	// Check if is inserted

	actualTranslateFile, err := repository.LoadTranslateFile(
		ctx,
		service.ID,
		expectedTranslateFile.Language,
	)
	if !assert.NoError(t, err) {
		return
	}

	assertEqualTranslateFile(t, expectedTranslateFile, actualTranslateFile)
}

func Test_SaveTranslateFileWithoutUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "Prepare test service") {
		return
	}

	// Actual Test

	expectedTranslateFile := randTranslateFile(randMessages())

	expectedTranslateFile.ID = uuid.Nil

	err = repository.SaveTranslateFile(ctx, service.ID, expectedTranslateFile)
	if !assert.NoError(t, err) {
		return
	}

	// Check if is inserted

	actualTranslateFile, err := repository.LoadTranslateFile(
		ctx,
		service.ID,
		expectedTranslateFile.Language,
	)
	if !assert.NoError(t, err) {
		return
	}

	assertEqualTranslateFile(t, expectedTranslateFile, actualTranslateFile)
}

func Test_UpdateTranslateFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "Prepare test service") {
		return
	}

	expectedTranslateFile := randTranslateFile(randMessages())

	err = repository.SaveTranslateFile(ctx, service.ID, expectedTranslateFile)
	if !assert.NoError(t, err) {
		return
	}

	// Actual Test

	expectedTranslateFile.Messages.Messages = randMessages()

	err = repository.SaveTranslateFile(ctx, service.ID, expectedTranslateFile)
	if !assert.NoError(t, err) {
		return
	}

	// Check if updated

	actualTranslateFile, err := repository.LoadTranslateFile(
		ctx,
		service.ID,
		expectedTranslateFile.Language,
	)
	if !assert.NoError(t, err) {
		return
	}

	assertEqualTranslateFile(t, expectedTranslateFile, actualTranslateFile)
}

func Test_SaveTranslateFileNoService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	translateFile := randTranslateFile(randMessages())

	err := repository.SaveTranslateFile(ctx, uuid.New(), translateFile)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func Test_LoadTranslateFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "Prepare test service") {
		return
	}

	expectedTranslateFile := randTranslateFile(randMessages())

	err = repository.SaveTranslateFile(ctx, service.ID, expectedTranslateFile)
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		expected    *model.TranslateFile
		expectedErr error
		name        string
		serviceID   uuid.UUID
	}{
		{
			name:        "All OK",
			expected:    expectedTranslateFile,
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

			actual, err := repository.LoadTranslateFile(
				ctx,
				tt.serviceID,
				expectedTranslateFile.Language,
			)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assertEqualTranslateFile(t, tt.expected, actual)
		})
	}
}
