package translate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"golang.org/x/text/language"
)

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.UploadTranslationFileRequest
		expectedErr error
		name        string
	}{
		{
			name: "Happy Path",
			input: &pb.UploadTranslationFileRequest{
				Language: "lv-LV",
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &pb.UploadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			expectedErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language tag",
			input: &pb.UploadTranslationFileRequest{
				Language: "",
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			expectedErr: errors.New("tag is not well-formed"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := parseUploadParams(tt.input)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.NotEmpty(t, parsed)
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
				Language: language.MustParse("lv-LV"),
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Empty data",
			input: uploadParams{
				Language: language.MustParse("lv-LV"),
				Schema:   pb.Schema_GO,
			},
			expectedErr: errors.New("'data' is required"),
		},
		{
			name: "Unspecified schema",
			input: uploadParams{
				Language: language.MustParse("lv-LV"),
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_UNSPECIFIED,
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

func Test_ParseDownloadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.DownloadTranslationFileRequest
		expectedErr error
		name        string
	}{
		{
			name: "Happy Path",
			input: &pb.DownloadTranslationFileRequest{
				Language: "lv-LV",
				Schema:   pb.Schema_GO,
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language tag",
			input: &pb.DownloadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Schema:   pb.Schema_GO,
			},
			expectedErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language",
			input: &pb.DownloadTranslationFileRequest{
				Language: "",
				Schema:   pb.Schema_GO,
			},
			expectedErr: errors.New("tag is not well-formed"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := parseDownloadParams(tt.input)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.NotEmpty(t, parsed)
		})
	}
}
