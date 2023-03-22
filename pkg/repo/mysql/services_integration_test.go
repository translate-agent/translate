package mysql

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/tracer"
)

var repository *Repo

func TestMain(m *testing.M) {
	ctx := context.Background()

	viper.SetEnvPrefix("translate_mysql")
	viper.AutomaticEnv()

	tp, err := tracer.TracerProvider()
	if err != nil {
		log.Panicf("set tracer provider: %v", err)
	}

	conf := &Conf{
		Host:     viper.GetString("host"),
		Port:     viper.GetString("port"),
		User:     viper.GetString("user"),
		Database: viper.GetString("database"),
	}

	repository, err = NewRepo(WithConf(ctx, conf))
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

	service := randService()
	err := repository.SaveService(context.Background(), service)

	assert.NoError(t, err)
}

func Test_UpdateService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "repository.SaveService method returned an error") {
		return
	}

	service.Name = gofakeit.FirstName()

	err = repository.SaveService(ctx, service)
	assert.NoError(t, err)
}

func Test_LoadService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "repository.SaveService method returned an error") {
		return
	}

	tests := []struct {
		expected    *model.Service
		expectedErr error
		name        string
		input       uuid.UUID
	}{
		{
			name:        "All OK",
			input:       service.ID,
			expected:    service,
			expectedErr: nil,
		},
		{
			name:        "Nonexistent",
			input:       uuid.New(),
			expectedErr: repo.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := repository.LoadService(ctx, tt.input)

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

	ctx := context.Background()

	expectedServices := make([]model.Service, 3)

	for i := 0; i < 3; i++ {
		service := randService()

		err := repository.SaveService(ctx, service)
		if !assert.NoError(t, err, "repository.SaveService method returned an error") {
			return
		}

		expectedServices[i] = *service
	}

	actual, err := repository.LoadServices(ctx)
	if !assert.NoError(t, err) {
		return
	}

	for _, expected := range expectedServices {
		if !assert.Contains(t, actual, expected) {
			return
		}
	}
}

func Test_DeleteService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	err := repository.SaveService(ctx, service)
	if !assert.NoError(t, err, "repository.SaveService method returned an error") {
		return
	}

	tests := []struct {
		expectedErr error
		name        string
		input       uuid.UUID
	}{
		{
			name:        "All OK",
			input:       service.ID,
			expectedErr: nil,
		},
		{
			name:        "Nonexistent",
			input:       uuid.New(),
			expectedErr: repo.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repository.DeleteService(ctx, tt.input)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
