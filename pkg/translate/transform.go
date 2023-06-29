package translate

import (
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

// ----------------------Common types----------------------

// uuidToProto converts uuid.UUID to string.
func uuidToProto(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}

	return u.String()
}

// uuidFromProto converts string to uuid.UUID.
func uuidFromProto(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}

	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse uuid: %w", err)
	}

	return id, nil
}

// languageToProto converts language.Tag to string.
func languageToProto(l language.Tag) string {
	return l.String()
}

// languageFromProto converts string to language.Tag.
func languageFromProto(s string) (language.Tag, error) {
	if s == "" {
		return language.Und, nil
	}

	l, err := language.Parse(s)
	if err != nil {
		return language.Und, fmt.Errorf("parse language: %w", err)
	}

	return l, nil
}

// sliceToProto converts a slice of type T to a slice of type *R using the provided elementToProto function.
func sliceToProto[T any, R any](slice []T, elementToProto func(*T) *R) []*R {
	if len(slice) == 0 {
		return nil
	}

	v := make([]*R, 0, len(slice))

	for i := range slice {
		v = append(v, elementToProto(&slice[i]))
	}

	return v
}

// sliceFromProto converts a slice of type *T to a slice of type R
// using the provided elementFromProto function.
func sliceFromProto[T any, R any](slice []*T, elementFromProto func(*T) (*R, error)) ([]R, error) {
	if len(slice) == 0 {
		return nil, nil
	}

	v := make([]R, 0, len(slice))

	for i := range slice {
		r, err := elementFromProto(slice[i])
		if err != nil {
			return nil, fmt.Errorf("transform element: %w", err)
		}

		v = append(v, *r)
	}

	return v, nil
}

// ----------------------Service----------------------

// serviceToProto converts model.Service to translatev1.Service.
func serviceToProto(s *model.Service) *translatev1.Service {
	if s == nil {
		return nil
	}

	return &translatev1.Service{Id: uuidToProto(s.ID), Name: s.Name}
}

// serviceFromProto converts translatev1.Service to model.Service.
func serviceFromProto(s *translatev1.Service) (*model.Service, error) {
	if s == nil {
		return nil, nil
	}

	var (
		service = &model.Service{Name: s.Name}
		err     error
	)

	service.ID, err = uuidFromProto(s.Id)
	if err != nil {
		return nil, fmt.Errorf("transform id: %w", err)
	}

	return service, nil
}

// servicesToProto converts []model.Service to []*translatev1.Service.
func servicesToProto(s []model.Service) []*translatev1.Service {
	return sliceToProto(s, serviceToProto)
}

// servicesFromProto converts []*translatev1.Service to []model.Service.
func servicesFromProto(s []*translatev1.Service) ([]model.Service, error) {
	return sliceFromProto(s, serviceFromProto)
}

// ----------------------Message----------------------

// messageToProto converts *model.Message to *translatev1.Message.
func messageToProto(m *model.Message) *translatev1.Message {
	if m == nil {
		return nil
	}

	return &translatev1.Message{
		Id:          m.ID,
		Message:     m.Message,
		Description: m.Description,
		Fuzzy:       m.Fuzzy,
	}
}

// messageSliceToProto converts []model.Message to []*translatev1.Message.
func messageSliceToProto(m []model.Message) []*translatev1.Message {
	return sliceToProto(m, messageToProto)
}

// ----------------------Messages----------------------

// messagesToProto converts *model.Messages to *translatev1.Messages.
func messagesToProto(m *model.Messages) *translatev1.Messages {
	if m == nil {
		return nil
	}

	return &translatev1.Messages{
		Language: languageToProto(m.Language),
		Messages: messageSliceToProto(m.Messages),
	}
}

// messagesSliceToProto converts []model.Messages to []*translatev1.Messages.
func messagesSliceToProto(m []model.Messages) []*translatev1.Messages {
	return sliceToProto(m, messagesToProto)
}
