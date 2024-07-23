package server

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil/expect"
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
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path With File ID",
			request: happyWithFileIDReq,
			wantErr: nil,
		},
		{
			name:    "Malformed language",
			request: malformedLangReq,
			wantErr: errors.New("parse language"),
		},

		{
			name:    "Malformed service ID",
			request: malformedServiceIDReq,
			wantErr: errors.New("parse service_id"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseUploadTranslationFileRequestParams(test.request)

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
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
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			params:  happyParams,
			wantErr: nil,
		},
		{
			name:    "Empty data",
			params:  emptyDataParams,
			wantErr: errors.New("'data' is required"),
		},
		{
			name:    "Unspecified schema",
			params:  unspecifiedSchemaParams,
			wantErr: errors.New("'schema' is required"),
		},
		{
			name:    "Unspecified service ID",
			params:  unspecifiedServiceIDParams,
			wantErr: errors.New("'service_id' is required"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
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
		wantErr error
		args    args
		want    language.Tag
		name    string
	}{
		{
			name:    "Translation language is defined/params undefined",
			args:    translationDefinedParamsUndefined,
			want:    translationDefinedParamsUndefined.translation.Language,
			wantErr: nil,
		},
		{
			name:    "Translation language is undefined/params defined",
			args:    translationUndefinedParamsDefined,
			want:    translationUndefinedParamsDefined.params.languageTag,
			wantErr: nil,
		},
		{
			name:    "Both defined, same language",
			args:    bothDefinedSameLang,
			want:    bothDefinedSameLang.translation.Language,
			wantErr: nil,
		},
		{
			name:    "Both undefined",
			args:    undefinedBoth,
			wantErr: errors.New("no language is set"),
		},
		{
			name:    "Language mismatch",
			args:    langMismatch,
			wantErr: errors.New("languages are mismatched"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := getLanguage(test.args.params, test.args.translation)

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if test.want != got {
				t.Errorf("want %s, got %s", test.want, got)
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
		wantErr error
		request *translatev1.DownloadTranslationFileRequest
		name    string
	}{
		{
			name:    "Happy Path",
			request: happyReq,
			wantErr: nil,
		},
		{
			name:    "Malformed service ID",
			request: malformedServiceIDReq,
			wantErr: errors.New("parse service_id"),
		},
		{
			name:    "Malformed language",
			request: malformedLangTagReq,
			wantErr: errors.New("parse language"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			params, err := parseDownloadTranslationFileRequestParams(test.request)

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
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
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			params:  happyParams,
			wantErr: nil,
		},
		{
			name:    "Unspecified schema",
			params:  unspecifiedSchemaParams,
			wantErr: errors.New("'schema' is required"),
		},
		{
			name:    "Unspecified service ID",
			params:  unspecifiedServiceIDParams,
			wantErr: errors.New("'service_id' is required"),
		},
		{
			name:    "Unspecified language",
			params:  unspecifiedLanguageTagReq,
			wantErr: errors.New("'language' is required"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
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
