package server

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

// -------------------Upload-----------------------

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.UploadTranslationFileRequest {
		return &translatev1.UploadTranslationFileRequest{
			Language:  gofakeit.LanguageBCP(),
			Data:      []byte(`{"key":"value"}`),
			Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
			ServiceId: gofakeit.UUID(),
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
			name:        "Malformed language tag",
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
			languageTag: language.MustParse(gofakeit.LanguageBCP()),
			data:        []byte(`{"key":"value"}`),
			schema:      translatev1.Schema(gofakeit.IntRange(1, 7)),
			serviceID:   uuid.New(),
		}
	}

	happyParams := randParams()

	emptyDataParams := randParams()
	emptyDataParams.data = nil

	unspecifiedSchemaParams := randParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedLangReq := randParams()
	unspecifiedLangReq.languageTag = language.Und

	unspecifiedServiceIDReq := randParams()
	unspecifiedServiceIDReq.serviceID = uuid.Nil

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
			name:        "Unspecified language",
			params:      unspecifiedLangReq,
			expectedErr: errors.New("'language' is required"),
		},
		{
			name:        "Unspecified service ID",
			params:      unspecifiedServiceIDReq,
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

// -------------------Download-----------------------

func Test_ParseDownloadParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.DownloadTranslationFileRequest {
		return &translatev1.DownloadTranslationFileRequest{
			Language:  gofakeit.LanguageBCP(),
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
			name:        "Malformed language tag",
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
			languageTag: language.MustParse(gofakeit.LanguageBCP()),
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
			name:        "Unspecified language tag",
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
