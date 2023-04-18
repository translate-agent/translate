package translate

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
		return nil, fmt.Errorf("parse language: %w", err)
	}

	params.serviceID, err = uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	params.translationFileID, err = uuidFromProto(req.GetTranslationFileId())
	if err != nil {
		return nil, fmt.Errorf("parse translation_file_id: %w", err)
	}

	return params, nil
}

func validateUploadTranslationFileRequestParams(params *uploadParams) error {
	if len(params.data) == 0 {
		return fmt.Errorf("'data' is required")
	}

	// Enforce that schema is present. (Temporal solution)
	if params.schema == translatev1.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	if params.serviceID == uuid.Nil {
		return fmt.Errorf("'service_id' is required")
	}

	if params.languageTag == language.Und {
		return fmt.Errorf("'language' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *translatev1.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	params, err := parseUploadTranslationFileRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = validateUploadTranslationFileRequestParams(params); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages, err := MessagesFromData(params.schema, params.data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Some converts do not provide language, so we override it with one from request for consistency.
	messages.Language = params.languageTag

	translationFile := &model.TranslationFile{
		ID:       params.translationFileID,
		Messages: messages,
	}

	switch err := t.repo.SaveTranslationFile(ctx, params.serviceID, translationFile); {
	default:
		return &emptypb.Empty{}, nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, "file not found")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "")
	}
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

	params.languageTag, err = langTagFromProto(req.GetLanguage())
	if err != nil {
		return nil, fmt.Errorf("parse language: %w", err)
	}

	return params, nil
}

func validateDownloadTranslationFileRequestParams(params *downloadParams) error {
	// Enforce that schema is present.
	if params.schema == translatev1.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	if params.serviceID == uuid.Nil {
		return fmt.Errorf("'service_id' is required")
	}

	if params.languageTag == language.Und {
		return fmt.Errorf("'language' is required")
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

	if err = validateDownloadTranslationFileRequestParams(params); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	translationFile, err := t.repo.LoadTranslationFile(ctx, params.serviceID, params.languageTag)

	switch {
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, "file not found")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "")
	}

	data, err := MessagesToData(params.schema, translationFile.Messages)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
