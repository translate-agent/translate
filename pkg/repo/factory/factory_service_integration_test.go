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
)

func Test_SaveService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subTest testutil.SubtestFn) {
		tests := []struct {
			service *model.Service
			name    string
		}{
			{
				name:    "With UUID",
				service: rand.ModelService(),
			},
			{
				name:    "Without UUID",
				service: rand.ModelService(rand.WithID(uuid.Nil)),
			},
		}
		for _, tt := range tests {
			tt := tt
			subTest(tt.name, func(ctx context.Context, t *testing.T) {
				err := repository.SaveService(ctx, tt.service)
				require.NoError(t, err, "Save service")

				// check if really saved
				actualService, err := repository.LoadService(ctx, tt.service.ID)
				require.NoError(t, err, "Load service saved service")

				assert.Equal(t, tt.service, actualService)
			})
		}
	})
}

func Test_UpdateService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		expectedService := rand.ModelService()

		err := repository.SaveService(testCtx, expectedService)
		require.NoError(t, err, "Prepare test service")

		// Actual Test

		// update service fields and save
		expectedService.Name = gofakeit.FirstName()

		err = repository.SaveService(testCtx, expectedService)
		require.NoError(t, err, "Update test service name")

		// check if really updated
		actualService, err := repository.LoadService(testCtx, expectedService.ID)
		require.NoError(t, err, "Load updated service")

		assert.Equal(t, expectedService, actualService)
	})
}

func Test_LoadService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)
		// Prepare
		service := rand.ModelService()

		err := repository.SaveService(testCtx, service)
		require.NoError(t, err, "Prepare test service")

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
				expected:    nil,
				expectedErr: repo.ErrNotFound,
			},
		}

		for _, tt := range tests {
			tt := tt
			subtest(tt.name, func(ctx context.Context, t *testing.T) {
				actual, err := repository.LoadService(ctx, tt.serviceID)
				require.ErrorIs(t, err, tt.expectedErr, "Load service")

				assert.Equal(t, tt.expected, actual)
			})
		}
	})
}

func Test_LoadServices(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		expectedServices := rand.ModelServices(3)
		for _, service := range expectedServices {
			err := repository.SaveService(testCtx, service)
			require.NoError(t, err, "Insert test service")
		}

		actual, err := repository.LoadServices(testCtx)
		require.NoError(t, err, "Load saved services")

		require.GreaterOrEqual(t, len(actual), len(expectedServices))

		for _, expected := range expectedServices {
			require.Contains(t, actual, *expected)
		}
	})
}

func Test_DeleteService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		service := rand.ModelService()

		err := repository.SaveService(testCtx, service)
		require.NoError(t, err, "Prepare test service")

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
				require.ErrorIs(t, err, tt.expectedErr, "Delete service")

				// check if really is deleted
				_, err = repository.LoadService(ctx, tt.serviceID)
				assert.ErrorIs(t, err, repo.ErrNotFound)
			})
		}
	})
}
