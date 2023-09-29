package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ----------------------UploadTranslationFile-------------------------------

type uploadParams struct {
	languageTag          language.Tag
	data                 []byte
	schema               translatev1.Schema
	serviceID            uuid.UUID
	original             bool
	populateTranslations bool
}

func parseUploadTranslationFileRequestParams(req *translatev1.UploadTranslationFileRequest) (*uploadParams, error) {
	var (
		params = &uploadParams{
			data:                 req.GetData(),
			schema:               req.GetSchema(),
			original:             req.GetOriginal(),
			populateTranslations: req.GetPopulateTranslations(),
		}
		err error
	)

	params.languageTag, err = languageFromProto(req.GetLanguage())
	if err != nil {
		return nil, fmt.Errorf("parse language: %w", err)
	}

	params.serviceID, err = uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	return params, nil
}

func (u *uploadParams) validate() error {
	if len(u.data) == 0 {
		return errors.New("'data' is required")
	}

	// Enforce that schema is present. (Temporal solution)
	if u.schema == translatev1.Schema_UNSPECIFIED {
		return errors.New("'schema' is required")
	}

	if u.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	return nil
}

// getLanguage returns the language tag for an upload based on the upload parameters and translation.
// It returns an error if no language is set or if the languages in the upload parameters and translation are mismatched.
func getLanguage(reqParams *uploadParams, translation *model.Translation) (language.Tag, error) {
	und := language.Und

	// Scenario 1: Both translation and params have undefined language
	if reqParams.languageTag == und && translation.Language == und {
		return und, errors.New("no language is set")
	}
	// Scenario 2: The languages in translation and params are different
	if reqParams.languageTag != und && translation.Language != und && translation.Language != reqParams.languageTag {
		return und, errors.New("languages are mismatched")
	}
	// Scenario 3: The language in translation is undefined but the language in params is defined
	if translation.Language == und {
		return reqParams.languageTag, nil
	}

	// Scenario 4 and 5:
	// The language in translation is defined but the language in params is undefined
	// The languages in translation and params are the same
	return translation.Language, nil
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *translatev1.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	params, err := parseUploadTranslationFileRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages, err := TranslationFromData(params)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages.Language, err = getLanguage(params, messages)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Case for when not original, or uploading original for the first time.
	updatedMessages := model.TranslationSlice{*messages}

	// When updating original translation, changes might affect translations - transform and update all translations.
	if messages.Original {
		all, err := t.repo.LoadTranslation(ctx, params.serviceID, repo.LoadTranslationOpts{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}

		if origIdx := all.OriginalIndex(); origIdx != -1 {
			oldOriginal := all[origIdx]

			// Mark new or altered translation as untranslated.
			all.MarkUntranslated(oldOriginal.FindChangedMessageIDs(messages))
			// Replace original translation with new ones.
			all.Replace(*messages)
			// Add missing translation for all translations.
			if params.populateTranslations {
				all.PopulateTranslations()
			}

			if err := t.fuzzyTranslate(ctx, all); err != nil {
				return nil, status.Errorf(codes.Internal, "")
			}

			updatedMessages = all
		}
	}

	for i := range updatedMessages {
		err = t.repo.SaveTranslation(ctx, params.serviceID, &updatedMessages[i])
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Errorf(codes.NotFound, "service not found")
		case err != nil:
			return nil, status.Errorf(codes.Internal, "")
		}
	}

	return &emptypb.Empty{}, nil
}

// ----------------------DownloadTranslationFile-------------------------------

type downloadParams struct {
	languageTag language.Tag
	schema      translatev1.Schema
	serviceID   uuid.UUID
}

func parseDownloadTranslationFileRequestParams(
	req *translatev1.DownloadTranslationFileRequest,
) (*downloadParams, error) {
	var (
		params = &downloadParams{schema: req.GetSchema()}
		err    error
	)

	params.serviceID, err = uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	params.languageTag, err = languageFromProto(req.GetLanguage())
	if err != nil {
		return nil, fmt.Errorf("parse language: %w", err)
	}

	return params, nil
}

func (d *downloadParams) validate() error {
	// Enforce that schema is present.
	if d.schema == translatev1.Schema_UNSPECIFIED {
		return errors.New("'schema' is required")
	}

	if d.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	if d.languageTag == language.Und {
		return errors.New("'language' is required")
	}

	return nil
}

func (t *TranslateServiceServer) DownloadTranslationFile(
	ctx context.Context,
	req *translatev1.DownloadTranslationFileRequest,
) (*translatev1.DownloadTranslationFileResponse, error) {
	params, err := parseDownloadTranslationFileRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages, err := t.repo.LoadTranslation(ctx, params.serviceID,
		repo.LoadTranslationOpts{FilterLanguages: []language.Tag{params.languageTag}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if len(messages) == 0 {
		messages = append(messages, model.Translation{Language: params.languageTag})
	}

	data, err := TranslationToData(params.schema, &messages[0])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
