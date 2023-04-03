package translate

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

// -------------------Upload-----------------------

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *translatev1.UploadTranslationFileRequest
		expectedErr error
		name        string
	}{
		{
			name: "Happy Path With File ID",
			input: &translatev1.UploadTranslationFileRequest{
				Language:        gofakeit.LanguageBCP(),
				Data:            []byte(`{"key":"value"}`),
				Schema:          translatev1.Schema(gofakeit.IntRange(1, 7)),
				ServiceId:       gofakeit.UUID(),
				TranslateFileId: gofakeit.UUID(),
			},
			expectedErr: nil,
		},
		{
			name: "Happy Path Without File ID",
			input: &translatev1.UploadTranslationFileRequest{
				Language:  gofakeit.LanguageBCP(),
				Data:      []byte(`{"key":"value"}`),
				Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
				ServiceId: gofakeit.UUID(),
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &translatev1.UploadTranslationFileRequest{
				Language: gofakeit.LanguageBCP() + "_FAIL",
			},
			expectedErr: errors.New("subtag \"fail\" is well-formed but unknown"),
		},
		{
			name: "Missing language tag",
			input: &translatev1.UploadTranslationFileRequest{
				Language: "",
			},
			expectedErr: errors.New("tag is not well-formed"),
		},
		{
			name: "Missing service ID",
			input: &translatev1.UploadTranslationFileRequest{
				Language:  gofakeit.LanguageBCP(),
				ServiceId: "",
			},
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name: "Malformed service ID",
			input: &translatev1.UploadTranslationFileRequest{
				Language:  gofakeit.LanguageBCP(),
				ServiceId: gofakeit.UUID() + "_FAIL",
			},
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name: "Malformed File ID",
			input: &translatev1.UploadTranslationFileRequest{
				Language:        gofakeit.LanguageBCP(),
				ServiceId:       gofakeit.UUID(),
				TranslateFileId: gofakeit.UUID() + "_FAIL",
			},
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*uploadTranslationFileRequest)(tt.input)

			params, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.NotEmpty(t, params)
		})
	}
}

func Test_ValidateUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr error
		input       uploadParams
	}{
		{
			name: "Happy Path",
			input: uploadParams{
				languageTag: language.MustParse(gofakeit.LanguageBCP()),
				data:        []byte(`{"key":"value"}`),
				schema:      translatev1.Schema(gofakeit.IntRange(1, 7)),
			},
			expectedErr: nil,
		},
		{
			name: "Empty data",
			input: uploadParams{
				languageTag: language.MustParse(gofakeit.LanguageBCP()),
				schema:      translatev1.Schema(gofakeit.IntRange(1, 7)),
			},
			expectedErr: errors.New("'data' is required"),
		},
		{
			name: "Unspecified schema",
			input: uploadParams{
				languageTag: language.MustParse(gofakeit.LanguageBCP()),
				data:        []byte(`{"key":"value"}`),
				schema:      translatev1.Schema_UNSPECIFIED,
			},
			expectedErr: errors.New("'schema' is required"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.validate()

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

	tests := []struct {
		expectedErr error
		input       *translatev1.DownloadTranslationFileRequest
		name        string
	}{
		{
			name: "Happy Path",
			input: &translatev1.DownloadTranslationFileRequest{
				ServiceId: gofakeit.UUID(),
				Language:  gofakeit.LanguageBCP(),
				Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
			},
			expectedErr: nil,
		},
		{
			name: "Missing service ID",
			input: &translatev1.DownloadTranslationFileRequest{
				ServiceId: "",
			},
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name: "Malformed service ID",
			input: &translatev1.DownloadTranslationFileRequest{
				ServiceId: gofakeit.UUID() + "_FAIL",
			},
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name: "Malformed language tag",
			input: &translatev1.DownloadTranslationFileRequest{
				ServiceId: gofakeit.UUID(),
				Language:  gofakeit.LanguageBCP() + "_FAIL",
			},
			expectedErr: errors.New("subtag \"fail\" is well-formed but unknown"),
		},
		{
			name: "Missing language",
			input: &translatev1.DownloadTranslationFileRequest{
				ServiceId: gofakeit.UUID(),
				Language:  "",
			},
			expectedErr: errors.New("tag is not well-formed"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*downloadTranslationFileRequest)(tt.input)

			params, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.NotEmpty(t, params)
		})
	}
}

func Test_ValidateDownloadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr error
		input       downloadParams
	}{
		{
			name: "Happy Path",
			input: downloadParams{
				languageTag: language.MustParse(gofakeit.LanguageBCP()),
				schema:      translatev1.Schema(gofakeit.IntRange(1, 7)),
			},
			expectedErr: nil,
		},
		{
			name: "Unspecified schema",
			input: downloadParams{
				languageTag: language.MustParse(gofakeit.LanguageBCP()),
				schema:      translatev1.Schema_UNSPECIFIED,
			},
			expectedErr: errors.New("'schema' is required"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.validate()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}
		})
	}
}
