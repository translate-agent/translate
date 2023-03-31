package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var ErrNotFound = errors.New("entity not found")

type ServicesRepo interface {
	// SaveService handles both Create and Update
	SaveService(ctx context.Context, service *model.Service) error
	LoadService(ctx context.Context, serviceID uuid.UUID) (*model.Service, error)
	LoadServices(ctx context.Context) ([]model.Service, error)
	DeleteService(ctx context.Context, serviceID uuid.UUID) error
}

type TranslateFileRepo interface {
	// SaveTranslateFile handles both Create and Update
	SaveTranslateFile(ctx context.Context, serviceID uuid.UUID, translateFile *model.TranslateFile) error
	LoadTranslateFile(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.TranslateFile, error)
}

type Repo interface {
	ServicesRepo
	TranslateFileRepo
}
