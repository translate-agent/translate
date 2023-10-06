package server

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// -------------------Upload-----------------------

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.UploadTranslationFileRequest {
		return &translatev1.UploadTranslationFileRequest{
			Language:  rand.Language().String(),
			Data:      []byte(`{"key":"value"}`),
			Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
			ServiceId: gofakeit.UUID(),
			Original:  gofakeit.Bool(),
		}
	}

	happyWithFileIDReq := randReq()

	malformedLangReq := randReq()
	malformedLangReq.Language += "_FAIL" //nolint:goconst

	malformedServiceIDReq := randReq()
	malformedServiceIDReq.ServiceId += "_FAIL"

	tests := []struct {
		request     *translatev1.UploadTranslationFileRequest
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path With File ID",
			request:     happyWithFileIDReq,
			expectedErr: nil,
		},
		{
			name:        "Malformed language",
			request:     malformedLangReq,
			expectedErr: errors.New("parse language"),
		},

		{
			name:        "Malformed service ID",
			request:     malformedServiceIDReq,
			expectedErr: errors.New("parse service_id"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseUploadTranslationFileRequestParams(tt.request)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.NotEmpty(t, params)
		})
	}
}

func Test_ValidateUploadParams(t *testing.T) {
	t.Parallel()

	randParams := func() *uploadParams {
		return &uploadParams{
			languageTag:          rand.Language(),
			data:                 []byte(`{"key":"value"}`),
			schema:               translatev1.Schema(gofakeit.IntRange(1, 7)),
			serviceID:            uuid.New(),
			original:             gofakeit.Bool(),
			populateTranslations: gofakeit.Bool(),
		}
	}

	happyParams := randParams()

	emptyDataParams := randParams()
	emptyDataParams.data = nil

	unspecifiedSchemaParams := randParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedServiceIDParams := randParams()
	unspecifiedServiceIDParams.serviceID = uuid.Nil

	tests := []struct {
		params      *uploadParams
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Empty data",
			params:      emptyDataParams,
			expectedErr: errors.New("'data' is required"),
		},
		{
			name:        "Unspecified schema",
			params:      unspecifiedSchemaParams,
			expectedErr: errors.New("'schema' is required"),
		},
		{
			name:        "Unspecified service ID",
			params:      unspecifiedServiceIDParams,
			expectedErr: errors.New("'service_id' is required"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.validate()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_GetLanguage(t *testing.T) {
	t.Parallel()

	type args struct {
		params   *uploadParams
		translation *model.Translation
	}

	// Tests

	translationDefinedParamsUndefined := args{
		params:   &uploadParams{languageTag: language.Und},
		translation: rand.ModelTranslation(3, nil),
	}

	translationUndefinedParamsDefined := args{
		params:   &uploadParams{languageTag: rand.Language()},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(language.Und)),
	}

	sameLang := rand.Language()
	bothDefinedSameLang := args{
		params:   &uploadParams{languageTag: sameLang},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(sameLang)),
	}

	undefinedBoth := args{
		params:   &uploadParams{languageTag: language.Und},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(language.Und)),
	}

	langs := rand.Languages(2)
	langMismatch := args{
		params:   &uploadParams{languageTag: langs[0]},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(langs[1])),
	}

	tests := []struct {
		expectedErr error
		args        args
		expected    language.Tag
		name        string
	}{
		{
			name:        "Translation language is defined/params undefined",
			args:        translationDefinedParamsUndefined,
			expected:    translationDefinedParamsUndefined.translation.Language,
			expectedErr: nil,
		},
		{
			name:        "Translation language is undefined/params defined",
			args:        translationUndefinedParamsDefined,
			expected:    translationUndefinedParamsDefined.params.languageTag,
			expectedErr: nil,
		},
		{
			name:        "Both defined, same language",
			args:        bothDefinedSameLang,
			expected:    bothDefinedSameLang.translation.Language,
			expectedErr: nil,
		},
		{
			name:        "Both undefined",
			args:        undefinedBoth,
			expectedErr: errors.New("no language is set"),
		},
		{
			name:        "Language mismatch",
			args:        langMismatch,
			expectedErr: errors.New("languages are mismatched"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := getLanguage(tt.args.params, tt.args.translation)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// -------------------Download-----------------------

func Test_ParseDownloadParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.DownloadTranslationFileRequest {
		return &translatev1.DownloadTranslationFileRequest{
			Language:  rand.Language().String(),
			Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
			ServiceId: gofakeit.UUID(),
		}
	}

	happyReq := randReq()

	missingServiceIDReq := randReq()
	missingServiceIDReq.ServiceId = ""

	malformedServiceIDReq := randReq()
	malformedServiceIDReq.ServiceId += "_FAIL"

	malformedLangTagReq := randReq()
	malformedLangTagReq.Language += "_FAIL"

	tests := []struct {
		expectedErr error
		request     *translatev1.DownloadTranslationFileRequest
		name        string
	}{
		{
			name:        "Happy Path",
			request:     happyReq,
			expectedErr: nil,
		},
		{
			name:        "Malformed service ID",
			request:     malformedServiceIDReq,
			expectedErr: errors.New("parse service_id"),
		},
		{
			name:        "Malformed language",
			request:     malformedLangTagReq,
			expectedErr: errors.New("parse language"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseDownloadTranslationFileRequestParams(tt.request)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.NotEmpty(t, params)
		})
	}
}

func Test_ValidateDownloadParams(t *testing.T) {
	t.Parallel()

	randParams := func() *downloadParams {
		return &downloadParams{
			languageTag: rand.Language(),
			schema:      translatev1.Schema(gofakeit.IntRange(1, 7)),
			serviceID:   uuid.New(),
		}
	}

	happyParams := randParams()

	unspecifiedSchemaParams := randParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedServiceIDParams := randParams()
	unspecifiedServiceIDParams.serviceID = uuid.Nil

	unspecifiedLanguageTagReq := randParams()
	unspecifiedLanguageTagReq.languageTag = language.Und

	tests := []struct {
		params      *downloadParams
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Unspecified schema",
			params:      unspecifiedSchemaParams,
			expectedErr: errors.New("'schema' is required"),
		},
		{
			name:        "Unspecified service ID",
			params:      unspecifiedServiceIDParams,
			expectedErr: errors.New("'service_id' is required"),
		},
		{
			name:        "Unspecified language",
			params:      unspecifiedLanguageTagReq,
			expectedErr: errors.New("'language' is required"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.validate()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
