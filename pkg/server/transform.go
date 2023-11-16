package server

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
		return nil, nil //nolint:nilnil
	}

	var (
		service = &model.Service{Name: s.GetName()}
		err     error
	)

	service.ID, err = uuidFromProto(s.GetId())
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
		Status:      translatev1.Message_Status(m.Status),
		Positions:   m.Positions,
	}
}

// messageFromProto converts *translatev1.Message to *model.Message.
func messageFromProto(m *translatev1.Message) (*model.Message, error) {
	if m == nil {
		return nil, nil //nolint:nilnil
	}

	return &model.Message{
		ID:          m.GetId(),
		Message:     m.GetMessage(),
		Description: m.GetDescription(),
		Status:      model.MessageStatus(m.GetStatus()),
		Positions:   m.GetPositions(),
	}, nil
}

// messagesToProto converts []model.Message to []*translatev1.Message.
func messagesToProto(m []model.Message) []*translatev1.Message {
	return sliceToProto(m, messageToProto)
}

// messagesFromProto converts []*translatev1.Message to []model.Message.
func messagesFromProto(m []*translatev1.Message) ([]model.Message, error) {
	return sliceFromProto(m, messageFromProto)
}

// ----------------------Translation----------------------

// translationToProto converts *model.Translation to *translatev1.Translation.
func translationToProto(t *model.Translation) *translatev1.Translation {
	if t == nil {
		return nil
	}

	return &translatev1.Translation{
		Language: languageToProto(t.Language),
		Original: t.Original,
		Messages: messagesToProto(t.Messages),
	}
}

// translationFromProto converts *translatev1.Translation to *model.Translation.
func translationFromProto(t *translatev1.Translation) (*model.Translation, error) {
	if t == nil {
		return nil, nil //nolint:nilnil
	}

	var (
		err         error
		translation = &model.Translation{Original: t.GetOriginal()}
	)

	if translation.Language, err = languageFromProto(t.GetLanguage()); err != nil {
		return nil, fmt.Errorf("transform language tag: %w", err)
	}

	if translation.Messages, err = messagesFromProto(t.GetMessages()); err != nil {
		return nil, fmt.Errorf("transform translation: %w", err)
	}

	return translation, nil
}

// translationsToProto converts []model.Translation to []*translatev1.Translation.
func translationsToProto(m []model.Translation) []*translatev1.Translation {
	return sliceToProto(m, translationToProto)
}

// ----------------------Mask----------------------

func maskFromProto(message proto.Message, mask *fieldmaskpb.FieldMask) (model.Mask, error) {
	if mask == nil {
		return nil, nil
	}

	paths := mask.GetPaths()

	if len(paths) == 0 {
		return model.Mask{}, nil
	}

	// Find corresponding model type to proto message.
	var modelType any

	switch message.(type) {
	case *translatev1.Message:
		modelType = new(model.Message)
	case *translatev1.Service:
		modelType = new(model.Service)
	case *translatev1.Translation:
		modelType = new(model.Translation)
	default:
		return nil, fmt.Errorf("unknown message type: %T", message)
	}

	var findFields func(reflect.Value, []string) []string
	findFields = func(v reflect.Value, currentPath []string) []string {
		var allFields []string

		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)

			allFields = append(allFields, strings.Join(append(currentPath, field.Name), "."))

			if field.Type.Kind() == reflect.Struct {
				findFields(v.Field(i), append(currentPath, field.Name))
			}
		}

		return allFields
	}

	allFields := findFields(reflect.ValueOf(modelType).Elem(), nil)

	modelMask := make(model.Mask, 0, len(paths))

	// Convert each proto path to a model path. e.g. "message" -> "Message", "plural_id" -> "PluralID"
	for _, path := range paths {
		idx := slices.IndexFunc(allFields, func(s string) bool {
			protoPath := strings.ToLower(strings.ReplaceAll(path, "_", ""))
			modelPath := strings.ToLower(s)

			return protoPath == modelPath
		})

		if idx != -1 {
			modelMask = append(modelMask, allFields[idx])
		} else {
			// should not happen, as the proto mask is checked before parsing it to model mask
			return nil, fmt.Errorf("unknown field: %s", path)
		}
	}

	return modelMask, nil
}
