package server

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// -------------------Upload-----------------------

func Test_ParseUploadParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.UploadTranslationFileRequest {
		return &translatev1.UploadTranslationFileRequest{
			Language:  rand.Language().String(),
			Data:      []byte(`{"key":"value"}`),
			Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
			ServiceId: gofakeit.UUID(),
			Original:  ptr(gofakeit.Bool()),
		}
	}

	happyWithFileIDReq := randReq()

	malformedLangReq := randReq()
	malformedLangReq.Language += "_FAIL" //nolint:goconst

	malformedServiceIDReq := randReq()
	malformedServiceIDReq.ServiceId += "_FAIL"

	tests := []struct {
		request *translatev1.UploadTranslationFileRequest
		wantErr string
		name    string
	}{
		{
			name:    "Happy Path With File ID",
			request: happyWithFileIDReq,
		},
		{
			name:    "Malformed language",
			request: malformedLangReq,
			wantErr: `parse language: parse language: language: subtag "fail" is well-formed but unknown`, //nolint:dupword
		},

		{
			name:    "Malformed service ID",
			request: malformedServiceIDReq,
			wantErr: "parse service_id: parse uuid: invalid UUID length: 41",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseUploadTranslationFileRequestParams(test.request)

			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if params == nil {
				t.Errorf("want params, got nil")
			}
		})
	}
}

func Test_ValidateUploadParams(t *testing.T) {
	t.Parallel()

	randParams := func() *uploadParams {
		return &uploadParams{
			languageTag:          rand.Language(),
			data:                 []byte(`{"key":"value"}`),
			schema:               translatev1.Schema(gofakeit.IntRange(1, 7)),
			serviceID:            uuid.New(),
			original:             ptr(gofakeit.Bool()),
			populateTranslations: gofakeit.Bool(),
		}
	}

	happyParams := randParams()

	emptyDataParams := randParams()
	emptyDataParams.data = nil

	unspecifiedSchemaParams := randParams()
	unspecifiedSchemaParams.schema = translatev1.Schema_UNSPECIFIED

	unspecifiedServiceIDParams := randParams()
	unspecifiedServiceIDParams.serviceID = uuid.Nil

	tests := []struct {
		params  *uploadParams
		wantErr string
		name    string
	}{
		{
			name:   "Happy Path",
			params: happyParams,
		},
		{
			name:    "Empty data",
			params:  emptyDataParams,
			wantErr: "'data' is required",
		},
		{
			name:    "Unspecified schema",
			params:  unspecifiedSchemaParams,
			wantErr: "'schema' is required",
		},
		{
			name:    "Unspecified service ID",
			params:  unspecifiedServiceIDParams,
			wantErr: "'service_id' is required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func Test_GetLanguage(t *testing.T) {
	t.Parallel()

	type args struct {
		params      *uploadParams
		translation *model.Translation
	}

	// Tests

	translationDefinedParamsUndefined := args{
		params:      &uploadParams{languageTag: language.Und},
		translation: rand.ModelTranslation(3, nil),
	}

	translationUndefinedParamsDefined := args{
		params:      &uploadParams{languageTag: rand.Language()},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(language.Und)),
	}

	sameLang := rand.Language()
	bothDefinedSameLang := args{
		params:      &uploadParams{languageTag: sameLang},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(sameLang)),
	}

	undefinedBoth := args{
		params:      &uploadParams{languageTag: language.Und},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(language.Und)),
	}

	langs := rand.Languages(2)
	langMismatch := args{
		params:      &uploadParams{languageTag: langs[0]},
		translation: rand.ModelTranslation(3, nil, rand.WithLanguage(langs[1])),
	}

	tests := []struct {
		wantErr string
		args    args
		want    language.Tag
		name    string
	}{
		{
			name: "Translation language is defined/params undefined",
			args: translationDefinedParamsUndefined,
			want: translationDefinedParamsUndefined.translation.Language,
		},
		{
			name: "Translation language is undefined/params defined",
			args: translationUndefinedParamsDefined,
			want: translationUndefinedParamsDefined.params.languageTag,
		},
		{
			name: "Both defined, same language",
			args: bothDefinedSameLang,
			want: bothDefinedSameLang.translation.Language,
		},
		{
			name:    "Both undefined",
			args:    undefinedBoth,
			wantErr: "no language is set",
		},
		{
			name:    "Language mismatch",
			args:    langMismatch,
			wantErr: "languages are mismatched",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := getLanguage(test.args.params, test.args.translation)

			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if test.want != got {
				t.Errorf("want language '%s', got '%s'", test.want, got)
			}
		})
	}
}

// -------------------Download-----------------------

func Test_ParseDownloadParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.DownloadTranslationFileRequest {
		return &translatev1.DownloadTranslationFileRequest{
			Language:  rand.Language().String(),
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
		wantErr string
		request *translatev1.DownloadTranslationFileRequest
		name    string
	}{
		{
			name:    "Happy Path",
			request: happyReq,
		},
		{
			name:    "Malformed service ID",
			request: malformedServiceIDReq,
			wantErr: "parse service_id: parse uuid: invalid UUID length: 41",
		},
		{
			name:    "Malformed language",
			request: malformedLangTagReq,
			wantErr: `parse language: parse language: language: subtag "fail" is well-formed but unknown`, //nolint:dupword
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseDownloadTranslationFileRequestParams(test.request)

			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if params == nil {
				t.Errorf("want params, got nil")
			}
		})
	}
}

func Test_ValidateDownloadParams(t *testing.T) {
	t.Parallel()

	randParams := func() *downloadParams {
		return &downloadParams{
			languageTag: rand.Language(),
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
		params  *downloadParams
		wantErr string
		name    string
	}{
		{
			name:   "Happy Path",
			params: happyParams,
		},
		{
			name:    "Unspecified schema",
			params:  unspecifiedSchemaParams,
			wantErr: "'schema' is required",
		},
		{
			name:    "Unspecified service ID",
			params:  unspecifiedServiceIDParams,
			wantErr: "'service_id' is required",
		},
		{
			name:    "Unspecified language",
			params:  unspecifiedLanguageTagReq,
			wantErr: "'language' is required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != "" {
				if err.Error() != test.wantErr {
					t.Errorf("\nwant '%s'\ngot  '%s'", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
			}
		})
	}
}

// helpers

// ptr returns pointer to the passed in value.
func ptr[T any](v T) *T {
	return &v
}
