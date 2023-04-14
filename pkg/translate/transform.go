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

// langTagToProto converts language.Tag to string.
func langTagToProto(l language.Tag) string {
	return l.String()
}

// langTagFromProto converts string to language.Tag.
func langTagFromProto(s string) (language.Tag, error) {
	// if s is empty string, language.Parse returns language.Und.
	l, err := language.Parse(s)
	if err != nil {
		return language.Und, fmt.Errorf("parse language tag: %w", err)
	}

	return l, nil
}

// ----------------------Service----------------------

// serviceToProto converts model.Service to translatev1.Service.
func serviceToProto(s *model.Service) *translatev1.Service {
	return &translatev1.Service{Id: uuidToProto(s.ID), Name: s.Name}
}

// serviceFromProto converts translatev1.Service to model.Service.
func serviceFromProto(s *translatev1.Service) (*model.Service, error) {
	var (
		service = &model.Service{Name: s.Name}
		err     error
	)

	service.ID, err = uuidFromProto(s.Id)
	if err != nil {
		return nil, fmt.Errorf("transform service id: %w", err)
	}

	return service, nil
}

// servicesToProto converts []model.Service to []*translatev1.Service.
func servicesToProto(s []model.Service) []*translatev1.Service {
	if len(s) == 0 {
		return nil
	}

	res := make([]*translatev1.Service, len(s))

	for i := range s {
		res[i] = serviceToProto(&s[i])
	}

	return res
}

// servicesFromProto converts []*translatev1.Service to []model.Service.
func servicesFromProto(s []*translatev1.Service) ([]model.Service, error) {
	if len(s) == 0 {
		return nil, nil
	}

	res := make([]model.Service, len(s))

	for i := range s {
		s, err := serviceFromProto(s[i])
		if err != nil {
			return nil, fmt.Errorf("transform service %d: %w", i, err)
		}

		res[i] = *s
	}

	return res, nil
}
