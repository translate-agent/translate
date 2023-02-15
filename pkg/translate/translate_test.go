package translate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateDownloadParams(t *testing.T) {
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
			name: "Bad formed language",
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
