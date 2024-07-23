//go:build integration

package factory

import (
	"context"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
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
		for _, test := range tests {
			subTest(test.name, func(ctx context.Context, t *testing.T) {
				err := repository.SaveService(ctx, test.service)
				if err != nil {
					t.Error(err)
					return
				}

				// check if really saved
				gotService, err := repository.LoadService(ctx, test.service.ID)
				if err != nil {
					t.Error(err)
					return
				}

				if !reflect.DeepEqual(test.service, gotService) {
					t.Errorf("want %v, got %v", test.service, gotService)
				}
			})
		}
	})
}

func Test_UpdateService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		wantService := rand.ModelService()

		err := repository.SaveService(testCtx, wantService)
		if err != nil {
			t.Error(err)
			return
		}

		// Test

		// update service fields and save
		wantService.Name = gofakeit.FirstName()

		err = repository.SaveService(testCtx, wantService)
		if err != nil {
			t.Error(err)
			return
		}

		// check if really updated
		gotService, err := repository.LoadService(testCtx, wantService.ID)
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(wantService, gotService) {
			t.Errorf("want %v, got %v", wantService, gotService)
		}
	})
}

func Test_LoadService(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)
		// Prepare
		service := rand.ModelService()

		err := repository.SaveService(testCtx, service)
		if err != nil {
			t.Error(err)
			return
		}

		tests := []struct {
			want      *model.Service
			wantErr   error
			name      string
			serviceID uuid.UUID
		}{
			{
				name:      "All OK",
				serviceID: service.ID,
				want:      service,
				wantErr:   nil,
			},
			{
				name:      "Not Found",
				serviceID: uuid.New(),
				want:      nil,
				wantErr:   repo.ErrNotFound,
			},
		}

		for _, test := range tests {
			subtest(test.name, func(ctx context.Context, t *testing.T) {
				got, err := repository.LoadService(ctx, test.serviceID)
				require.ErrorIs(t, err, test.wantErr, "Load service")

				if !reflect.DeepEqual(test.want, got) {
					t.Errorf("want %v, got %v", test.want, got)
				}
			})
		}
	})
}

func Test_LoadServices(t *testing.T) {
	t.Parallel()

	allRepos(t, func(t *testing.T, repository repo.Repo, subtest testutil.SubtestFn) {
		testCtx, _ := testutil.Trace(t)

		// Prepare
		wantServices := rand.ModelServices(3)
		for _, service := range wantServices {
			err := repository.SaveService(testCtx, service)
			if err != nil {
				t.Error(err)
				return
			}
		}

		got, err := repository.LoadServices(testCtx)
		if err != nil {
			t.Error(err)
			return
		}

		require.GreaterOrEqual(t, len(got), len(wantServices))

		for _, want := range wantServices {
			require.Contains(t, got, *want)
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
		if err != nil {
			t.Error(err)
			return
		}

		tests := []struct {
			wantErr   error
			name      string
			serviceID uuid.UUID
		}{
			{
				name:      "All OK",
				serviceID: service.ID,
				wantErr:   nil,
			},
			{
				name:      "Nonexistent",
				serviceID: uuid.New(),
				wantErr:   repo.ErrNotFound,
			},
		}

		for _, test := range tests {
			subtest(test.name, func(ctx context.Context, t *testing.T) {
				err := repository.DeleteService(ctx, test.serviceID)
				require.ErrorIs(t, err, test.wantErr, "Delete service")

				// check if really is deleted
				_, err = repository.LoadService(ctx, test.serviceID)
				require.ErrorIs(t, err, repo.ErrNotFound)
			})
		}
	})
}
