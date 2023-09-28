package repo

import (
	"context"
	"errors"
	"io"

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

type LoadTranslationOpts struct {
	FilterLanguages []language.Tag
}

type TranslationRepo interface {
	// SaveTranslation handles both Create and Update
	SaveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error
	LoadTranslation(ctx context.Context, serviceID uuid.UUID, opts LoadTranslationOpts) (model.TranslationSlice, error)
}

type Repo interface {
	ServicesRepo
	TranslationRepo

	io.Closer
}
