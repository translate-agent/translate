package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
)

var ErrNotFound = errors.New("entity not found")

type ServicesRepo interface {
	SaveService(ctx context.Context, service *model.Service) error // handles both create and update
	LoadService(ctx context.Context, serviceID uuid.UUID) (*model.Service, error)
	LoadServices(ctx context.Context) ([]model.Service, error)
	DeleteService(ctx context.Context, serviceID uuid.UUID) error
}

type TranslateFileRepo interface {
	SaveTranslateFile(ctx context.Context, serviceID uuid.UUID, translateFile *model.TranslateFile) error
	LoadTranslateFile(ctx context.Context, serviceID uuid.UUID, language model.Language) (*model.TranslateFile, error)
}

type Repo interface {
	ServicesRepo
	TranslateFileRepo
}
