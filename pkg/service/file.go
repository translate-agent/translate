package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	if u.languageTag == language.Und {
		return errors.New("'language' is required")
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

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages, err := MessagesFromData(params.schema, params.data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Some converts do not provide language, so we override it with one from request for consistency.
	messages.Language = params.languageTag

	switch err := t.repo.SaveMessages(ctx, params.serviceID, &messages); {
	default:
		return &emptypb.Empty{}, nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, "service not found")
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

	messages, err := t.repo.LoadMessages(ctx, params.serviceID, params.languageTag)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	data, err := MessagesToData(params.schema, *messages)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
