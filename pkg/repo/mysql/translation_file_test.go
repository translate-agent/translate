package mysql

import (
	"encoding/json"
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
	"golang.org/x/text/language"
)

func Test_TransformTranslationFile(t *testing.T) {
	t.Parallel()

	conf := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, _ *rand.Rand) {
			values[0] = reflect.ValueOf(&model.TranslationFile{
				ID: uuid.New(),
				Messages: model.Messages{
					Language: randDbLanguageTag().Tag,
					Messages: randDbMessageSlice(),
				},
			})
		},
	}

	f := func(expectedTranslationFile *model.TranslationFile) bool {
		restoredTranslationFile := toTranslationFile(fromTranslationFile(expectedTranslationFile))
		return assert.Equal(t, expectedTranslationFile, restoredTranslationFile)
	}

	assert.NoError(t, quick.Check(f, conf))
}

func randDbMessageSlice() dbMessageSlice {
	msgCount := gofakeit.IntRange(1, 10)
	messages := make([]model.Message, msgCount)

	for i := 0; i < msgCount; i++ {
		messages[i] = model.Message{
			ID:          gofakeit.Sentence(10),
			Message:     gofakeit.Sentence(10),
			Description: gofakeit.Sentence(10),
			Fuzzy:       gofakeit.Bool(),
		}
	}

	return messages
}

func dbMessageSliceToBytes(t *testing.T, messages dbMessageSlice) []byte {
	bytes, err := json.Marshal(messages)
	require.NoError(t, err)

	return bytes
}

func Test_ScanDbMessageSlice(t *testing.T) {
	t.Parallel()

	happyPath := dbMessageSliceToBytes(t, randDbMessageSlice())

	unsupportedType := "unsupported type"

	randomBytes := gofakeit.ImagePng(10, 10)

	tests := []struct {
		expectedErr error
		src         interface{}
		name        string
	}{
		{
			name:        "Happy path",
			src:         happyPath,
			expectedErr: nil,
		},
		{
			name:        "Unsupported type",
			src:         unsupportedType,
			expectedErr: errors.New("unsupported type"),
		},
		{
			name:        "Random bytes",
			src:         randomBytes,
			expectedErr: errors.New("unmarshal messages"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var dbSlice dbMessageSlice
			err := dbSlice.Scan(tt.src)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_ValueDbMessageSlice(t *testing.T) {
	t.Parallel()

	dbSlice := randDbMessageSlice()

	_, err := dbSlice.Value()
	assert.NoError(t, err)
}

func randDbLanguageTag() dbLanguageTag {
	return dbLanguageTag{Tag: language.MustParse(gofakeit.LanguageBCP())}
}

func randDbLanguageTagToBytes(langTag dbLanguageTag) []uint8 {
	return []uint8(langTag.String())
}

func Test_ScanDbLanguageTag(t *testing.T) {
	t.Parallel()

	happyPath := randDbLanguageTagToBytes(randDbLanguageTag())

	unsupportedType := "unsupported type"

	randomBytes := gofakeit.ImagePng(10, 10)

	tests := []struct {
		expectedErr error
		src         interface{}
		name        string
	}{
		{
			name:        "Happy path",
			src:         happyPath,
			expectedErr: nil,
		},
		{
			name:        "Unsupported type",
			src:         unsupportedType,
			expectedErr: errors.New("unsupported type"),
		},
		{
			name:        "Random bytes",
			src:         randomBytes,
			expectedErr: errors.New("parse language"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var dbTag dbLanguageTag
			err := dbTag.Scan(tt.src)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_ValueDbLanguageTag(t *testing.T) {
	t.Parallel()

	dbTag := randDbLanguageTag()

	_, err := dbTag.Value()
	assert.NoError(t, err)
}
