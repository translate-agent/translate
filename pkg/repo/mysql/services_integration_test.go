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
	"go.expect.digital/translate/pkg/testutil"
)

func randService() *model.Service {
	return &model.Service{
		Name: gofakeit.FirstName(),
		ID:   uuid.New(),
	}
}

func Test_SaveService(t *testing.T) {
	t.Parallel()

	_, subtest := testutil.Trace(t)

	tests := []struct {
		service *model.Service
		name    string
	}{
		{
			name:    "With UUID",
			service: randService(),
		},
		{
			name:    "Without UUID",
			service: &model.Service{Name: gofakeit.Name()},
		},
	}
	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			err := repository.SaveService(ctx, tt.service)
			if !assert.NoError(t, err) {
				return
			}

			// check if really saved
			actualService, err := repository.LoadService(ctx, tt.service.ID)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.service, actualService)
		})
	}
}

func Test_UpdateService(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
	expectedService := randService()

	err := repository.SaveService(ctx, expectedService)
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	// Actual Test

	// update service fields and save
	expectedService.Name = gofakeit.FirstName()

	err = repository.SaveService(ctx, expectedService)
	if !assert.NoError(t, err) {
		return
	}

	// check if really updated
	actualService, err := repository.LoadService(ctx, expectedService.ID)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expectedService, actualService)
}

func Test_LoadService(t *testing.T) {
	t.Parallel()

	testCtx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	err := repository.SaveService(testCtx, service)
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	tests := []struct {
		expected    *model.Service
		expectedErr error
		name        string
		serviceID   uuid.UUID
	}{
		{
			name:        "All OK",
			serviceID:   service.ID,
			expected:    service,
			expectedErr: nil,
		},
		{
			name:        "Not Found",
			serviceID:   uuid.New(),
			expectedErr: repo.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			actual, err := repository.LoadService(ctx, tt.serviceID)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_LoadServices(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
	expectedServices := make([]model.Service, 3)

	for i := 0; i < 3; i++ {
		service := randService()

		err := repository.SaveService(ctx, service)
		if !assert.NoError(t, err, "Prepare test data") {
			return
		}

		expectedServices[i] = *service
	}

	actual, err := repository.LoadServices(ctx)
	if !assert.NoError(t, err) {
		return
	}

	require.GreaterOrEqual(t, len(actual), len(expectedServices))

	for _, expected := range expectedServices {
		if !assert.Contains(t, actual, expected) {
			return
		}
	}
}

func Test_DeleteService(t *testing.T) {
	t.Parallel()

	testCtx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	err := repository.SaveService(testCtx, service)
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	tests := []struct {
		expectedErr error
		name        string
		serviceID   uuid.UUID
	}{
		{
			name:        "All OK",
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
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			err := repository.DeleteService(ctx, tt.serviceID)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			// check if really is deleted
			_, err = repository.LoadService(ctx, tt.serviceID)
			assert.ErrorIs(t, err, repo.ErrNotFound)
		})
	}
}
