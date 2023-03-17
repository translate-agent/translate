package mysql

import (
	"context"
	"fmt"
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

var mysqlRepo *Repo

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

	mysqlRepo, err = NewRepo(WithDBConfig(ctx, conf))
	if err != nil {
		log.Panicf("create new repo: %v", err)
	}

	code := m.Run()

	mysqlRepo.db.Close()

	if err := tp.Shutdown(ctx); err != nil {
		log.Panicf("tp shutdown: %v", err)
	}

	os.Exit(code)
}

func createTestService() *model.Service {
	return &model.Service{
		Name: gofakeit.FirstName(),
		ID:   uuid.New(),
	}
}

// Inserts service to db.
func insertTestService(t *testing.T, service *model.Service) error {
	t.Helper()

	_, err := mysqlRepo.db.Exec(`INSERT INTO service (id, name) VALUES (?, ?)`, service.ID, service.Name)
	if err != nil {
		return fmt.Errorf("insert test service: %w", err)
	}

	return nil
}

func Test_MysqlSaveService(t *testing.T) {
	t.Parallel()

	service := createTestService()
	err := mysqlRepo.SaveService(context.Background(), service)

	assert.NoError(t, err)
}

func Test_MysqlUpdateService(t *testing.T) {
	t.Parallel()

	service := createTestService()

	err := insertTestService(t, service)
	if !assert.NoError(t, err) {
		return
	}

	service.Name = gofakeit.FirstName()

	err = mysqlRepo.SaveService(context.Background(), service)
	assert.NoError(t, err)
}

func Test_MysqlLoadService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := createTestService()

	err := insertTestService(t, service)
	if !assert.NoError(t, err) {
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

			actual, err := mysqlRepo.LoadService(ctx, tt.input)

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

func Test_MysqlLoadServices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedServices := make([]model.Service, 3)

	for i := 0; i < 3; i++ {
		service := createTestService()

		err := insertTestService(t, service)
		if !assert.NoError(t, err) {
			return
		}

		expectedServices[i] = *service
	}

	actual, err := mysqlRepo.LoadServices(ctx)
	if !assert.NoError(t, err) {
		return
	}

	for _, expected := range expectedServices {
		if !assert.Contains(t, actual, expected) {
			return
		}
	}
}

func Test_MysqlDeleteService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := createTestService()

	err := insertTestService(t, service)
	if !assert.NoError(t, err) {
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

			err := mysqlRepo.DeleteService(ctx, tt.input)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
