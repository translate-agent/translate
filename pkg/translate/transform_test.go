package translate

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_TransformUUID(t *testing.T) {
	t.Parallel()

	t.Run("UUID to string to UUID", func(t *testing.T) {
		t.Parallel()

		f := func(expectedID uuid.UUID) bool {
			restoredID, err := uuidFromProto(uuidToProto(expectedID))

			return assert.NoError(t, err) && assert.Equal(t, expectedID, restoredID)
		}

		assert.NoError(t, quick.Check(f, &quick.Config{MaxCount: 1000}))
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
		restoredLangTag, err := languagesFromProto(languagesToProto(expectedLangTag))

		return assert.NoError(t, err) && assert.Equal(t, expectedLangTag, restoredLangTag)
	}

	assert.NoError(t, quick.Check(f, conf))
}

func Test_TransformService(t *testing.T) {
	t.Parallel()

	t.Run("Service to proto to Service", func(t *testing.T) {
		t.Parallel()

		f := func(expectedService model.Service) bool {
			restoredService, err := serviceFromProto(serviceToProto(&expectedService))

			return assert.NoError(t, err) && assert.Equal(t, expectedService, *restoredService)
		}

		assert.NoError(t, quick.Check(f, &quick.Config{MaxCount: 1000}))
	})

	t.Run("Services to proto to services", func(t *testing.T) {
		t.Parallel()

		f := func(expectedServices []model.Service) bool {
			restoredServices, err := servicesFromProto(servicesToProto(expectedServices))

			return assert.NoError(t, err) && assert.ElementsMatch(t, expectedServices, restoredServices)
		}
		assert.NoError(t, quick.Check(f, &quick.Config{MaxCount: 100}))
	})
}
