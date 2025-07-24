package server

import (
	"errors"
	"math/rand"
	"reflect"
	"slices"
	"testing"
	"testing/quick"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Test_TransformUUID(t *testing.T) {
	t.Parallel()

	t.Run("UUID to string to UUID", func(t *testing.T) {
		t.Parallel()

		f := func(wantID uuid.UUID) bool {
			restoredID, err := uuidFromProto(uuidToProto(wantID))
			if err != nil {
				t.Error(err)
				return false
			}

			return wantID == restoredID
		}

		err := quick.Check(f, &quick.Config{MaxCount: 1000})
		if err != nil {
			t.Error(err)
		}
	})

	// Separate check with Nil UUID.
	t.Run("Nil UUID to string to UUID", func(t *testing.T) {
		t.Parallel()

		wantID := uuid.Nil

		restoredID, err := uuidFromProto(uuidToProto(wantID))
		if err != nil {
			t.Error(err)
			return
		}

		if wantID != restoredID {
			t.Errorf("want UUID %s, got %s", wantID, restoredID)
		}
	})
}

func Test_TransformLanguage(t *testing.T) {
	t.Parallel()

	conf := &quick.Config{
		MaxCount: 1000,
		Values: func(values []reflect.Value, _ *rand.Rand) {
			values[0] = reflect.ValueOf(language.MustParse(gofakeit.LanguageBCP()))
		},
	}

	f := func(wantLangTag language.Tag) bool {
		restoredLangTag, err := languageFromProto(languageToProto(wantLangTag))
		if err != nil {
			t.Error(err)
			return false
		}

		if wantLangTag != restoredLangTag {
			t.Errorf("want language '%s', got '%s'", wantLangTag, restoredLangTag)
			return false
		}

		return true
	}

	err := quick.Check(f, conf)
	if err != nil {
		t.Error(err)
	}
}

func Test_TransformService(t *testing.T) {
	t.Parallel()

	t.Run("Service to proto to Service", func(t *testing.T) {
		t.Parallel()

		f := func(wantService model.Service) bool {
			restoredService, err := serviceFromProto(serviceToProto(&wantService))
			if err != nil {
				t.Error(err)
				return false
			}

			return wantService == *restoredService
		}

		err := quick.Check(f, &quick.Config{MaxCount: 1000})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Services to proto to services", func(t *testing.T) {
		t.Parallel()

		f := func(wantServices []model.Service) bool {
			restoredServices, err := servicesFromProto(servicesToProto(wantServices))
			if err != nil {
				t.Error(err)
				return false
			}

			if len(wantServices) != 0 && len(restoredServices) != 0 && !reflect.DeepEqual(wantServices, restoredServices) {
				t.Logf("\nwant %v\ngot  %v", wantServices, restoredServices)
				return false
			}

			return true
		}

		err := quick.Check(f, &quick.Config{MaxCount: 100})
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_maskFromProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		protoMessage proto.Message          // proto message type
		protoMask    *fieldmaskpb.FieldMask // mask as received from the request
		wantErr      error
		modelMask    model.Mask // parsed mask with correct model paths
	}{
		// positive tests
		{
			name:         "nil mask",
			protoMessage: new(translatev1.Service),
			protoMask:    nil,
			modelMask:    nil,
		},
		{
			name:         "service mask",
			protoMessage: new(translatev1.Service), // corresponds to model.Service
			protoMask:    &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			modelMask:    model.Mask{"name"},
		},
		{
			name:         "message mask",
			protoMessage: new(translatev1.Message), // corresponds to model.Message
			protoMask:    &fieldmaskpb.FieldMask{Paths: []string{"plural_id", "message", "description", "status", "positions"}},
			modelMask:    model.Mask{"pluralid", "message", "description", "status", "positions"},
		},
		{
			name:         "translation mask",
			protoMessage: new(translatev1.Translation), // corresponds to model.Translation
			protoMask:    &fieldmaskpb.FieldMask{Paths: []string{"language", "original", "messages"}},
			modelMask:    model.Mask{"language", "original", "messages"},
		},
		// negative tests
		{
			name:         "empty mask",
			protoMessage: new(translatev1.Service),
			protoMask:    &fieldmaskpb.FieldMask{},
			modelMask:    model.Mask{},
			wantErr:      errors.New("field mask must contain at least 1 path"),
		},
		{
			name:         "mask not nil, message nil",
			protoMessage: nil,
			protoMask:    &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			modelMask:    nil,
			wantErr:      errors.New("message cannot be nil"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := maskFromProto(test.protoMessage, test.protoMask)
			if test.wantErr != nil {
				if test.wantErr.Error() != err.Error() {
					t.Errorf("want error '%s', got '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if !slices.Equal(test.modelMask, got) {
				t.Errorf("want mask %v, got %v", test.modelMask, got)
			}
		})
	}
}
