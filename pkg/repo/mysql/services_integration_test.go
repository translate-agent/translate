//go:build integration

package mysql

import (
	"context"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/tracer"
)

var repository *Repo

func TestMain(m *testing.M) {
	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	tp, err := tracer.TracerProvider()
	if err != nil {
		log.Panicf("set tracer provider: %v", err)
	}

	repository, err = NewRepo(WithDefaultDB(ctx))
	if err != nil {
		log.Panicf("create new repo: %v", err)
	}

	code := m.Run()

	repository.db.Close()

	if err := tp.Shutdown(ctx); err != nil {
		log.Panicf("tp shutdown: %v", err)
	}

	os.Exit(code)
}

func randService() *model.Service {
	return &model.Service{
		Name: gofakeit.FirstName(),
		ID:   uuid.New(),
	}
}

func Test_SaveService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repository.SaveService(ctx, tt.service)
			require.NoError(t, err)

			// check if really saved
			actualService, err := repository.LoadService(ctx, tt.service.ID)
			require.NoError(t, err)

			assert.Equal(t, tt.service, actualService)
		})
	}
}

func Test_UpdateService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedService := randService()

	err := repository.SaveService(ctx, expectedService)
	require.NoError(t, err, "Prepare test service")

	// update service fields and save
	expectedService.Name = gofakeit.FirstName()

	err = repository.SaveService(ctx, expectedService)
	require.NoError(t, err, "Update service")

	// check if really updated
	actualService, err := repository.LoadService(ctx, expectedService.ID)
	require.NoError(t, err, "Load updated service")

	assert.Equal(t, expectedService, actualService)
}

func Test_LoadService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
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
			name:        "Nonexistent",
			serviceID:   uuid.New(),
			expectedErr: &repo.NotFoundError{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := repository.LoadService(ctx, tt.serviceID)

			if tt.expectedErr != nil {
				e := reflect.New(reflect.TypeOf(tt.expectedErr).Elem()).Interface()
				assert.ErrorAs(t, err, &e)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_LoadServices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	count := gofakeit.IntRange(1, 7)

	expectedServices := make([]model.Service, count)

	for i := 0; i < count; i++ {
		service := randService()

		err := repository.SaveService(ctx, service)
		require.NoError(t, err, "Prepare test service")

		expectedServices[i] = *service
	}

	actual, err := repository.LoadServices(ctx)
	require.NoError(t, err)

	for _, expected := range expectedServices {
		require.Contains(t, actual, expected)
	}
}

func Test_DeleteService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
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
			expectedErr: &repo.DefaultError{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repository.DeleteService(ctx, tt.serviceID)
			if tt.expectedErr != nil {
				e := reflect.New(reflect.TypeOf(tt.expectedErr).Elem()).Interface()
				assert.ErrorAs(t, err, &e)

				return
			}

			require.NoError(t, err)

			// check if really is deleted
			_, err = repository.LoadService(ctx, tt.serviceID)
			var e *repo.NotFoundError
			assert.ErrorAs(t, err, &e)
		})
	}
}
