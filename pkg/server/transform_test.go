package server

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		f := func(expectedID uuid.UUID) bool {
			restoredID, err := uuidFromProto(uuidToProto(expectedID))
			require.NoError(t, err)

			return assert.Equal(t, expectedID, restoredID)
		}

		require.NoError(t, quick.Check(f, &quick.Config{MaxCount: 1000}))
	})

	// Separate check with Nil UUID.
	t.Run("Nil UUID to string to UUID", func(t *testing.T) {
		t.Parallel()

		expectedID := uuid.Nil

		restoredID, err := uuidFromProto(uuidToProto(expectedID))
		require.NoError(t, err)

		assert.Equal(t, expectedID, restoredID)
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

	f := func(expectedLangTag language.Tag) bool {
		restoredLangTag, err := languageFromProto(languageToProto(expectedLangTag))
		require.NoError(t, err)

		return assert.Equal(t, expectedLangTag, restoredLangTag)
	}

	require.NoError(t, quick.Check(f, conf))
}

func Test_TransformService(t *testing.T) {
	t.Parallel()

	t.Run("Service to proto to Service", func(t *testing.T) {
		t.Parallel()

		f := func(expectedService model.Service) bool {
			restoredService, err := serviceFromProto(serviceToProto(&expectedService))
			require.NoError(t, err)

			return assert.Equal(t, expectedService, *restoredService)
		}

		require.NoError(t, quick.Check(f, &quick.Config{MaxCount: 1000}))
	})

	t.Run("Services to proto to services", func(t *testing.T) {
		t.Parallel()

		f := func(expectedServices []model.Service) bool {
			restoredServices, err := servicesFromProto(servicesToProto(expectedServices))
			require.NoError(t, err)

			return assert.ElementsMatch(t, expectedServices, restoredServices)
		}

		require.NoError(t, quick.Check(f, &quick.Config{MaxCount: 100}))
	})
}

func Test_maskFromProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		protoMessage proto.Message          // proto message type
		protoMask    *fieldmaskpb.FieldMask // mask as received from the request
		expectedErr  error
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
			expectedErr:  errors.New("field mask must contain at least 1 path"),
		},
		{
			name:         "mask not nil, message nil",
			protoMessage: nil,
			protoMask:    &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			modelMask:    nil,
			expectedErr:  errors.New("message cannot be nil"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := maskFromProto(tt.protoMessage, tt.protoMask)
			if tt.expectedErr != nil {
				require.EqualError(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.modelMask, actual)
		})
	}
}
