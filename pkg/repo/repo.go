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

type TranslationFileRepo interface {
	// SaveTranslationFile handles both Create and Update
	SaveTranslationFile(ctx context.Context, serviceID uuid.UUID, translationFile *model.TranslationFile) error
	LoadTranslationFile(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.TranslationFile, error)
}

type Repo interface {
	ServicesRepo
	TranslationFileRepo
}
