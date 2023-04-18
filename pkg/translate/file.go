package translate

import (
	"context"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ----------------------UploadTranslationFile-------------------------------

type uploadParams struct {
	languageTag       language.Tag
	data              []byte
	schema            translatev1.Schema
	serviceID         uuid.UUID
	translationFileID uuid.UUID
}

func parseUploadTranslationFileRequestParams(req *translatev1.UploadTranslationFileRequest) (*uploadParams, error) {
	var (
		params = &uploadParams{data: req.GetData(), schema: req.GetSchema()}
		err    error
	)

	params.languageTag, err = langTagFromProto(req.GetLanguage())
	if err != nil {
		return nil, &fieldViolationError{field: "language", err: err}
	}

	params.serviceID, err = uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, &fieldViolationError{field: "service_id", err: err}
	}

	params.translationFileID, err = uuidFromProto(req.GetTranslationFileId())
	if err != nil {
		return nil, &fieldViolationError{field: "translation_file_id", err: err}
	}

	return params, nil
}

func validateUploadTranslationFileRequestParams(params *uploadParams) error {
	if len(params.data) == 0 {
		return &fieldViolationError{field: "data", err: errEmptyField}
	}

	// Enforce that schema is present. (Temporal solution)
	if params.schema == translatev1.Schema_UNSPECIFIED {
		return &fieldViolationError{field: "schema", err: errEmptyField}
	}

	if params.serviceID == uuid.Nil {
		return &fieldViolationError{field: "service_id", err: errEmptyField}
	}

	if params.languageTag == language.Und {
		return &fieldViolationError{field: "language", err: errEmptyField}
	}

	return nil
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *translatev1.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	params, err := parseUploadTranslationFileRequestParams(req)
	if err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err = validateUploadTranslationFileRequestParams(params); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	messages, err := MessagesFromData(params.schema, params.data)
	if err != nil {
		return nil, convertFromErrorToStatusErr(&convertError{field: "data", err: err, schema: params.schema.String()})
	}

	// Some converts do not provide language, so we override it with one from request for consistency.
	messages.Language = params.languageTag

	translationFile := &model.TranslationFile{
		ID:       params.translationFileID,
		Messages: messages,
	}

	if err := t.repo.SaveTranslationFile(ctx, params.serviceID, translationFile); err != nil {
		return nil, repoErrorToStatusErr(err)
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
		return nil, &fieldViolationError{field: "service_id", err: err}
	}

	params.languageTag, err = langTagFromProto(req.GetLanguage())
	if err != nil {
		return nil, &fieldViolationError{field: "language", err: err}
	}

	return params, nil
}

func validateDownloadTranslationFileRequestParams(params *downloadParams) error {
	// Enforce that schema is present.
	if params.schema == translatev1.Schema_UNSPECIFIED {
		return &fieldViolationError{field: "schema", err: errEmptyField}
	}

	if params.serviceID == uuid.Nil {
		return &fieldViolationError{field: "service_id", err: errEmptyField}
	}

	if params.languageTag == language.Und {
		return &fieldViolationError{field: "language", err: errEmptyField}
	}

	return nil
}

func (t *TranslateServiceServer) DownloadTranslationFile(
	ctx context.Context,
	req *translatev1.DownloadTranslationFileRequest,
) (*translatev1.DownloadTranslationFileResponse, error) {
	params, err := parseDownloadTranslationFileRequestParams(req)
	if err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err = validateDownloadTranslationFileRequestParams(params); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	translationFile, err := t.repo.LoadTranslationFile(ctx, params.serviceID, params.languageTag)
	if err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	data, err := MessagesToData(params.schema, translationFile.Messages)
	if err != nil {
		return nil, convertToErrorToStatusErr(&convertError{err: err, schema: params.schema.String()})
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
