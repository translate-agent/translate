package mysql

import (
	"context"
	"log"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
)

var mysqlRepo *Repo

func init() {
	viper.SetEnvPrefix("translate_mysql")
	viper.AutomaticEnv()

	mysqlConf := Conf{
		Addr:     viper.GetString("address"),
		User:     viper.GetString("user"),
		Database: viper.GetString("database"),
	}

	db, err := NewDB(context.Background(), &mysqlConf)
	if err != nil {
		log.Panicf("create new db: %v", err)
	}

	insertQuery := `REPLACE INTO Service (ID, Name)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Service 1'
  ),
  (
    '11111111-1111-1111-1111-111111111111',
    'Service 2'
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'Service 3'
  );`

	_, err = db.ExecContext(context.Background(), insertQuery)
	if err != nil {
		log.Panicf("insert mock services: %v", err)
	}

	mysqlRepo, err = NewRepo(WithDB(db))
	if err != nil {
		log.Panicf("create new repo: %v", err)
	}
}

func Test_MysqlSaveService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		input *model.Service
		name  string
	}{
		{
			name: "Save Service",
			input: &model.Service{
				ID:   uuid.MustParse(gofakeit.UUID()),
				Name: gofakeit.FirstName(),
			},
		},
		{
			name: "Update Service",
			input: &model.Service{
				ID:   uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				Name: gofakeit.FirstName(),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := mysqlRepo.SaveService(ctx, tt.input)
			assert.NoError(t, err)
		})
	}
}

func Test_MysqlLoadService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		expected    *model.Service
		expectedErr error
		name        string
		input       uuid.UUID
	}{
		{
			name:  "All OK",
			input: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			expected: &model.Service{
				Name: "Service 1",
				ID:   uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			},
			expectedErr: nil,
		},
		{
			name:        "Not exists",
			input:       uuid.MustParse("99999999-9999-9999-9999-999999999999"),
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

	tests := []struct {
		expectedErr  error
		name         string
		minimalCount int
	}{
		{
			name:         "All OK",
			minimalCount: 3,
			expectedErr:  nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			services, err := mysqlRepo.LoadServices(ctx)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			actualCount := len(services)
			assert.GreaterOrEqual(t, actualCount, tt.minimalCount)
		})
	}
}

func Test_MysqlDeleteService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		expectedErr error
		name        string
		input       uuid.UUID
	}{
		{
			name:        "All OK",
			input:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			expectedErr: nil,
		},
		{
			name:        "Not exists",
			input:       uuid.MustParse("99999999-9999-9999-9999-999999999999"),
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
