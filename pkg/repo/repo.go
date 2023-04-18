package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

type NotFoundError struct {
	Fields map[string]string
	Entity string
}

func (n *NotFoundError) Error() string {
	fields := make([]string, 0, len(n.Fields))

	for k, v := range n.Fields {
		fields = append(fields, fmt.Sprintf("%s '%s'", k, v))
	}

	return fmt.Sprintf("%s with %s does not exist", n.Entity, strings.Join(fields, " and "))
}

type DefaultError struct {
	Entity    string
	Err       error
	Operation string
}

func (d *DefaultError) Error() string {
	return fmt.Sprintf("%s %s: %s", d.Operation, d.Entity, d.Err.Error())
}

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
