package translate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

// -------------------Upload-----------------------

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr error
		input       *translatev1.UploadTranslationFileRequest
		expected    uploadParams
	}{
		{
			name: "Happy Path",
			input: &translatev1.UploadTranslationFileRequest{
				Language: "lv",
				Data:     []byte(`{"key":"value"}`),
				Schema:   translatev1.Schema_GO,
			},
			expected: uploadParams{
				language: language.Latvian,
				data:     []byte(`{"key":"value"}`),
				schema:   translatev1.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &translatev1.UploadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Data:     []byte(`{"key":"value"}`),
				Schema:   translatev1.Schema_GO,
			},
			expectedErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language tag",
			input: &translatev1.UploadTranslationFileRequest{
				Language: "",
				Data:     []byte(`{"key":"value"}`),
				Schema:   translatev1.Schema_GO,
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

			req := (*uploadTranslationFileRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, actual)
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
				language: language.MustParse("lv-LV"),
				data:     []byte(`{"key":"value"}`),
				schema:   translatev1.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Empty data",
			input: uploadParams{
				language: language.MustParse("lv-LV"),
				schema:   translatev1.Schema_GO,
			},
			expectedErr: errors.New("'data' is required"),
		},
		{
			name: "Unspecified schema",
			input: uploadParams{
				language: language.MustParse("lv-LV"),
				data:     []byte(`{"key":"value"}`),
				schema:   translatev1.Schema_UNSPECIFIED,
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
		name        string
		input       *translatev1.DownloadTranslationFileRequest
		expectedErr error
		expected    downloadParams
	}{
		{
			name: "Happy Path",
			input: &translatev1.DownloadTranslationFileRequest{
				Language: "lv",
				Schema:   translatev1.Schema_GO,
			},
			expected: downloadParams{
				language: language.Latvian,
				schema:   translatev1.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &translatev1.DownloadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Schema:   translatev1.Schema_GO,
			},
			expectedErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language",
			input: &translatev1.DownloadTranslationFileRequest{
				Language: "",
				Schema:   translatev1.Schema_GO,
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

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, actual)
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
				language: language.MustParse("lv-LV"),
				schema:   translatev1.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Unspecified schema",
			input: downloadParams{
				language: language.MustParse("lv-LV"),
				schema:   translatev1.Schema_UNSPECIFIED,
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
