package translate

import (
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
			Language:          gofakeit.LanguageBCP(),
			Data:              []byte(`{"key":"value"}`),
			Schema:            translatev1.Schema(gofakeit.IntRange(1, 7)),
			ServiceId:         gofakeit.UUID(),
			TranslationFileId: gofakeit.UUID(),
		}
	}

	happyWithFileIDReq := randReq()

	happyWithoutFileIDReq := randReq()
	happyWithoutFileIDReq.TranslationFileId = ""

	malformedLangReq := randReq()
	malformedLangReq.Language += "_FAIL" //nolint:goconst

	malformedServiceIDReq := randReq()
	malformedServiceIDReq.ServiceId += "_FAIL"

	malformedFileIDReq := randReq()
	malformedFileIDReq.TranslationFileId += "_FAIL"

	tests := []struct {
		request     *translatev1.UploadTranslationFileRequest
		expectedErr *parseParamError
		name        string
	}{
		{
			name:        "Happy Path With File ID",
			request:     happyWithFileIDReq,
			expectedErr: nil,
		},
		{
			name:        "Happy Path Without File ID",
			request:     happyWithoutFileIDReq,
			expectedErr: nil,
		},
		{
			name:        "Malformed language tag",
			request:     malformedLangReq,
			expectedErr: &parseParamError{field: "language"},
		},

		{
			name:        "Malformed service ID",
			request:     malformedServiceIDReq,
			expectedErr: &parseParamError{field: "service_id"},
		},
		{
			name:        "Malformed File ID",
			request:     malformedFileIDReq,
			expectedErr: &parseParamError{field: "translation_file_id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseUploadTranslationFileRequestParams(tt.request)

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
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
			languageTag:       language.MustParse(gofakeit.LanguageBCP()),
			data:              []byte(`{"key":"value"}`),
			schema:            translatev1.Schema(gofakeit.IntRange(1, 7)),
			serviceID:         uuid.New(),
			translationFileID: uuid.New(),
		}
	}

	happyParams := randParams()

	emptyDataParams := randParams()
	emptyDataParams.data = nil

	unspecifiedSchemaParams := randParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedLangParams := randParams()
	unspecifiedLangParams.languageTag = language.Und

	unspecifiedServiceParams := randParams()
	unspecifiedServiceParams.serviceID = uuid.Nil

	tests := []struct {
		params      *uploadParams
		name        string
		expectedErr *validateParamError
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Empty data",
			params:      emptyDataParams,
			expectedErr: &validateParamError{param: "data"},
		},
		{
			name:        "Unspecified schema",
			params:      unspecifiedSchemaParams,
			expectedErr: &validateParamError{param: "schema"},
		},
		{
			name:        "Unspecified language",
			params:      unspecifiedLangParams,
			expectedErr: &validateParamError{param: "language"},
		},
		{
			name:        "Unspecified service ID",
			params:      unspecifiedServiceParams,
			expectedErr: &validateParamError{param: "service_id"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateUploadTranslationFileRequestParams(tt.params)

			if tt.expectedErr != nil {
				var e *validateParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.param, e.param)
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
		expectedErr *parseParamError
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
			expectedErr: &parseParamError{field: "service_id"},
		},
		{
			name:        "Malformed language tag",
			request:     malformedLangTagReq,
			expectedErr: &parseParamError{field: "language"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseDownloadTranslationFileRequestParams(tt.request)

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
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

	unspecifiedLangParams := randParams()
	unspecifiedLangParams.languageTag = language.Und

	tests := []struct {
		params      *downloadParams
		expectedErr *validateParamError
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
			expectedErr: &validateParamError{param: "schema"},
		},
		{
			name:        "Unspecified service ID",
			params:      unspecifiedServiceIDParams,
			expectedErr: &validateParamError{param: "service_id"},
		},
		{
			name:        "Unspecified language tag",
			params:      unspecifiedLangParams,
			expectedErr: &validateParamError{param: "language"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateDownloadTranslationFileRequestParams(tt.params)

			if tt.expectedErr != nil {
				var e *validateParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.param, e.param)
				return
			}

			assert.NoError(t, err)
		})
	}
}
