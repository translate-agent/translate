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
		req     *pb.UploadTranslationFileRequest
		wantErr error
		name    string
	}{
		{
			name: "Happy Path",
			req: &pb.UploadTranslationFileRequest{
				Language: "lv-LV",
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			req: &pb.UploadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			wantErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language tag",
			req: &pb.UploadTranslationFileRequest{
				Language: "",
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			wantErr: errors.New("tag is not well-formed"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := parseUploadParams(tt.req)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, parsed)
		})
	}
}

func Test_ValidateUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr error
		params  uploadParams
	}{
		{
			name: "Happy Path",
			params: uploadParams{
				Language: language.MustParse("lv-LV"),
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_GO,
			},
			wantErr: nil,
		},
		{
			name: "Empty data",
			params: uploadParams{
				Language: language.MustParse("lv-LV"),
				Schema:   pb.Schema_GO,
			},
			wantErr: errors.New("'data' is required"),
		},
		{
			name: "Unspecified schema",
			params: uploadParams{
				Language: language.MustParse("lv-LV"),
				Data:     []byte(`{"key":"value"}`),
				Schema:   pb.Schema_UNSPECIFIED,
			},
			wantErr: errors.New("'schema' must be specified"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.validate()

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_ParseDownloadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		req     *pb.DownloadTranslationFileRequest
		wantErr error
		name    string
	}{
		{
			name: "Happy Path",
			req: &pb.DownloadTranslationFileRequest{
				Language: "lv-LV",
				Schema:   pb.Schema_GO,
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			req: &pb.DownloadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Schema:   pb.Schema_GO,
			},
			wantErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language",
			req: &pb.DownloadTranslationFileRequest{
				Language: "",
				Schema:   pb.Schema_GO,
			},
			wantErr: errors.New("tag is not well-formed"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := parseDownloadParams(tt.req)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, parsed)
		})
	}
}
