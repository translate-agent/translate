package translate

import (
	"context"

	"github.com/google/uuid"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ----------------------UploadTranslationFile-------------------------------

type uploadParams struct {
	languageTag language.Tag
	data        []byte
	schema      translatev1.Schema
	serviceID   uuid.UUID
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

	return params, nil
}

func (u *uploadParams) validate() error {
	if len(u.data) == 0 {
		return &fieldViolationError{field: "data", err: errEmptyField}
	}

	// Enforce that schema is present. (Temporal solution)
	if u.schema == translatev1.Schema_UNSPECIFIED {
		return &fieldViolationError{field: "schema", err: errEmptyField}
	}

	if u.serviceID == uuid.Nil {
		return &fieldViolationError{field: "service_id", err: errEmptyField}
	}

	if u.languageTag == language.Und {
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

	if err = params.validate(); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	messages, err := MessagesFromData(params.schema, params.data)
	if err != nil {
		return nil, convertFromErrorToStatusErr(&convertError{field: "data", err: err, schema: params.schema.String()})
	}

	// Some converts do not provide language, so we override it with one from request for consistency.
	messages.Language = params.languageTag

	if err := t.repo.SaveMessages(ctx, params.serviceID, &messages); err != nil {
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

func (d *downloadParams) validate() error {
	// Enforce that schema is present.
	if d.schema == translatev1.Schema_UNSPECIFIED {
		return &fieldViolationError{field: "schema", err: errEmptyField}
	}

	if d.serviceID == uuid.Nil {
		return &fieldViolationError{field: "service_id", err: errEmptyField}
	}

	if d.languageTag == language.Und {
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

	if err = params.validate(); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	messages, err := t.repo.LoadMessages(ctx, params.serviceID, params.languageTag)
	if err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	data, err := MessagesToData(params.schema, *messages)
	if err != nil {
		return nil, convertToErrorToStatusErr(&convertError{err: err, schema: params.schema.String()})
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
