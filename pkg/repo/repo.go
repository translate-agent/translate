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

type LoadMessagesOpts struct {
	FilterLanguages []language.Tag
}

type MessagesRepo interface {
	// SaveMessages handles both Create and Update
	SaveMessages(ctx context.Context, serviceID uuid.UUID, messages *model.Messages) error
	LoadMessages(ctx context.Context, serviceID uuid.UUID, opts LoadMessagesOpts) ([]model.Messages, error)
}

type Repo interface {
	ServicesRepo
	MessagesRepo
}
