package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo/common"
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

// getLanguage returns the language tag for an upload based on the upload parameters and messages.
// It returns an error if no language is set or if the languages in the upload parameters and messages are mismatched.
func getLanguage(reqParams *uploadParams, messages *model.Messages) (language.Tag, error) {
	und := language.Und

	// Scenario 1: Both messages and params have undefined language
	if reqParams.languageTag == und && messages.Language == und {
		return und, errors.New("no language is set")
	}
	// Scenario 2: The languages in messages and params are different
	if reqParams.languageTag != und && messages.Language != und && messages.Language != reqParams.languageTag {
		return und, errors.New("languages are mismatched")
	}
	// Scenario 3: The language in messages is undefined but the language in params is defined
	if messages.Language == und {
		return reqParams.languageTag, nil
	}

	// Scenario 4 and 5:
	// The language in messages is defined but the language in params is undefined
	// The languages in messages and params are the same
	return messages.Language, nil
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

	messages, err := MessagesFromData(params.schema, params.data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// If after converting original is false, override it with the value from the request.
	// For now, only XLIFF formats can determine if the file is original or not.
	// All other format's converts marks this flag as false.
	if !messages.Original {
		messages.Original = params.original
	}

	messages.Language, err = getLanguage(params, messages)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	allMessages, err := t.repo.LoadMessages(ctx, params.serviceID, common.LoadMessagesOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	switch err := t.repo.SaveMessages(ctx, params.serviceID, messages); {
	default:
		// noop
	case errors.Is(err, common.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, "service not found")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "")
	}

	if messages.Original && len(allMessages) != 0 {
		// TODO: optimize performance when populateTranslations param is true, currently saveMessages() will be called twice.

		// find original messages where text has been altered then translate & update associated messages for all translations.
		if err := t.updateAlteredMessages(ctx, params.serviceID, allMessages, messages); err != nil {
			return nil, err
		}

		// if populateTranslations flag is true, populate the translated messages.
		if params.populateTranslations {
			if err = t.populateTranslatedMessages(ctx, params.serviceID, messages, allMessages); err != nil {
				return nil, err
			}
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

	messages, err := t.repo.LoadMessages(ctx, params.serviceID,
		common.LoadMessagesOpts{FilterLanguages: []language.Tag{params.languageTag}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if len(messages) == 0 {
		messages = append(messages, model.Messages{Language: params.languageTag})
	}

	data, err := MessagesToData(params.schema, &messages[0])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
