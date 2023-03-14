package translate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	tpb "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

// -------------------Upload-----------------------

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr error
		input       *tpb.UploadTranslationFileRequest
		expected    uploadParams
	}{
		{
			name: "Happy Path",
			input: &tpb.UploadTranslationFileRequest{
				Language: "lv",
				Data:     []byte(`{"key":"value"}`),
				Schema:   tpb.Schema_GO,
			},
			expected: uploadParams{
				language: language.Latvian,
				data:     []byte(`{"key":"value"}`),
				schema:   tpb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &tpb.UploadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Data:     []byte(`{"key":"value"}`),
				Schema:   tpb.Schema_GO,
			},
			expectedErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language tag",
			input: &tpb.UploadTranslationFileRequest{
				Language: "",
				Data:     []byte(`{"key":"value"}`),
				Schema:   tpb.Schema_GO,
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
				schema:   tpb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Empty data",
			input: uploadParams{
				language: language.MustParse("lv-LV"),
				schema:   tpb.Schema_GO,
			},
			expectedErr: errors.New("'data' is required"),
		},
		{
			name: "Unspecified schema",
			input: uploadParams{
				language: language.MustParse("lv-LV"),
				data:     []byte(`{"key":"value"}`),
				schema:   tpb.Schema_UNSPECIFIED,
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
		input       *tpb.DownloadTranslationFileRequest
		expectedErr error
		expected    downloadParams
	}{
		{
			name: "Happy Path",
			input: &tpb.DownloadTranslationFileRequest{
				Language: "lv",
				Schema:   tpb.Schema_GO,
			},
			expected: downloadParams{
				language: language.Latvian,
				schema:   tpb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &tpb.DownloadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Schema:   tpb.Schema_GO,
			},
			expectedErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language",
			input: &tpb.DownloadTranslationFileRequest{
				Language: "",
				Schema:   tpb.Schema_GO,
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
				schema:   tpb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Unspecified schema",
			input: downloadParams{
				language: language.MustParse("lv-LV"),
				schema:   tpb.Schema_UNSPECIFIED,
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
