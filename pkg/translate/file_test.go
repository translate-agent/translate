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

func randUploadReq() *translatev1.UploadTranslationFileRequest {
	return &translatev1.UploadTranslationFileRequest{
		Language:          gofakeit.LanguageBCP(),
		Data:              []byte(`{"key":"value"}`),
		Schema:            translatev1.Schema(gofakeit.IntRange(1, 7)),
		ServiceId:         gofakeit.UUID(),
		TranslationFileId: gofakeit.UUID(),
	}
}

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	happyWithFileIDReq := randUploadReq()

	happyWithoutFileIDReq := randUploadReq()
	happyWithoutFileIDReq.TranslationFileId = ""

	malformedLangReq := randUploadReq()
	malformedLangReq.Language += "_FAIL" //nolint:goconst

	malformedServiceIDReq := randUploadReq()
	malformedServiceIDReq.ServiceId += "_FAIL"

	malformedFileIDReq := randUploadReq()
	malformedFileIDReq.TranslationFileId += "_FAIL"

	tests := []struct {
		input       *translatev1.UploadTranslationFileRequest
		expectedErr *parseParamError
		name        string
	}{
		{
			name:        "Happy Path With File ID",
			input:       happyWithFileIDReq,
			expectedErr: nil,
		},
		{
			name:        "Happy Path Without File ID",
			input:       happyWithoutFileIDReq,
			expectedErr: nil,
		},
		{
			name:        "Malformed language tag",
			input:       malformedLangReq,
			expectedErr: &parseParamError{field: "language"},
		},

		{
			name:        "Malformed service ID",
			input:       malformedServiceIDReq,
			expectedErr: &parseParamError{field: "service_id"},
		},
		{
			name:        "Malformed File ID",
			input:       malformedFileIDReq,
			expectedErr: &parseParamError{field: "translation_file_id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*uploadTranslationFileRequest)(tt.input)

			params, err := req.parseParams()

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

func randUploadParams() uploadParams {
	return uploadParams{
		languageTag:       language.MustParse(gofakeit.LanguageBCP()),
		data:              []byte(`{"key":"value"}`),
		schema:            translatev1.Schema(gofakeit.IntRange(1, 7)),
		serviceID:         uuid.New(),
		translationFileID: uuid.New(),
	}
}

func Test_ValidateUploadParams(t *testing.T) {
	t.Parallel()

	happyParams := randUploadParams()

	emptyDataParams := randUploadParams()
	emptyDataParams.data = nil

	unspecifiedSchemaParams := randUploadParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedLangParams := randUploadParams()
	unspecifiedLangParams.languageTag = language.Und

	unspecifiedServiceParams := randUploadParams()
	unspecifiedServiceParams.serviceID = uuid.Nil

	tests := []struct {
		name        string
		expectedErr *validateParamError
		input       uploadParams
	}{
		{
			name:        "Happy Path",
			input:       happyParams,
			expectedErr: nil,
		},
		{
			name:        "Empty data",
			input:       emptyDataParams,
			expectedErr: &validateParamError{param: "data"},
		},
		{
			name:        "Unspecified schema",
			input:       unspecifiedSchemaParams,
			expectedErr: &validateParamError{param: "schema"},
		},
		{
			name:        "Unspecified language",
			input:       unspecifiedLangParams,
			expectedErr: &validateParamError{param: "language"},
		},
		{
			name:        "Unspecified service ID",
			input:       unspecifiedServiceParams,
			expectedErr: &validateParamError{param: "service_id"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.validate()

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

func randDownloadReq() *translatev1.DownloadTranslationFileRequest {
	return &translatev1.DownloadTranslationFileRequest{
		Language:  gofakeit.LanguageBCP(),
		Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
		ServiceId: gofakeit.UUID(),
	}
}

func Test_ParseDownloadParams(t *testing.T) {
	t.Parallel()

	happyReq := randDownloadReq()

	missingServiceIDReq := randDownloadReq()
	missingServiceIDReq.ServiceId = ""

	malformedServiceIDReq := randDownloadReq()
	malformedServiceIDReq.ServiceId += "_FAIL"

	malformedLangTagReq := randDownloadReq()
	malformedLangTagReq.Language += "_FAIL"

	tests := []struct {
		expectedErr *parseParamError
		input       *translatev1.DownloadTranslationFileRequest
		name        string
	}{
		{
			name:        "Happy Path",
			input:       happyReq,
			expectedErr: nil,
		},
		{
			name:        "Malformed service ID",
			input:       malformedServiceIDReq,
			expectedErr: &parseParamError{field: "service_id"},
		},
		{
			name:        "Malformed language tag",
			input:       malformedLangTagReq,
			expectedErr: &parseParamError{field: "language"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*downloadTranslationFileRequest)(tt.input)

			params, err := req.parseParams()

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

func randDownloadParams() downloadParams {
	return downloadParams{
		languageTag: language.MustParse(gofakeit.LanguageBCP()),
		schema:      translatev1.Schema(gofakeit.IntRange(1, 7)),
		serviceID:   uuid.New(),
	}
}

func Test_ValidateDownloadParams(t *testing.T) {
	t.Parallel()

	happyParams := randDownloadParams()

	unspecifiedSchemaParams := randDownloadParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedServiceIDParams := randDownloadParams()
	unspecifiedServiceIDParams.serviceID = uuid.Nil

	unspecifiedLangParams := randDownloadParams()
	unspecifiedLangParams.languageTag = language.Und

	tests := []struct {
		name        string
		expectedErr *validateParamError
		input       downloadParams
	}{
		{
			name:        "Happy Path",
			input:       happyParams,
			expectedErr: nil,
		},
		{
			name:        "Unspecified schema",
			input:       unspecifiedSchemaParams,
			expectedErr: &validateParamError{param: "schema"},
		},
		{
			name:        "Unspecified service ID",
			input:       unspecifiedServiceIDParams,
			expectedErr: &validateParamError{param: "service_id"},
		},
		{
			name:        "Unspecified language tag",
			input:       unspecifiedLangParams,
			expectedErr: &validateParamError{param: "language"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.validate()

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
