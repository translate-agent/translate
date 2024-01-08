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

type LoadTranslationsOpts struct {
	FilterLanguages []language.Tag
}

type TranslationsRepo interface {
	// SaveTranslation handles both Create and Update
	SaveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error
	LoadTranslations(ctx context.Context, serviceID uuid.UUID, opts LoadTranslationsOpts) (model.Translations, error)
}

type Repo interface {
	ServicesRepo
	TranslationsRepo

	io.Closer

	// Tx executes fn in a transaction.
	Tx(ctx context.Context, fn func(ctx context.Context, repo Repo) error) error
}
