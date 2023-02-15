package translate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateUploadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr error
		params  UploadParams
	}{
		{
			name: "Happy Path",
			params: UploadParams{
				Language: LanguageData{Str: "lv-LV"},
				Data:     []byte(`{"key":"value"}`),
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			params: UploadParams{
				Language: LanguageData{Str: "xyz-ZY-Latn"},
				Data:     []byte(`{"key":"value"}`),
			},
			wantErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name: "Missing language",
			params: UploadParams{
				Data: []byte(`{"key":"value"}`),
			},
			wantErr: errors.New("'language' is required"),
		},
		{
			name: "Empty data",
			params: UploadParams{
				Language: LanguageData{Str: "lv-LV"},
			},
			wantErr: errors.New("'data' is required"),
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
			assert.NotEmpty(t, tt.params.Language.Tag)
		})
	}
}

func Test_ValidateDownloadParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr error
		params  DownloadParams
	}{
		{
			name: "Happy Path",
			params: DownloadParams{
				Language: LanguageData{Str: "lv-LV"},
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			params: DownloadParams{
				Language: LanguageData{Str: "xyz-ZY-Latn"},
			},
			wantErr: errors.New("subtag \"xyz\" is well-formed but unknown"),
		},
		{
			name:    "Missing language",
			params:  DownloadParams{},
			wantErr: errors.New("'language' is required"),
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
			assert.NotEmpty(t, tt.params.Language.Tag)
		})
	}
}
