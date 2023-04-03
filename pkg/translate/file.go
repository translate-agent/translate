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

type (
	uploadTranslationFileRequest   translatev1.UploadTranslationFileRequest
	downloadTranslationFileRequest translatev1.DownloadTranslationFileRequest
)

// ----------------------UploadTranslationFile-------------------------------

type uploadParams struct {
	languageTag       language.Tag
	data              []byte
	schema            translatev1.Schema
	serviceID         uuid.UUID
	translationFileID uuid.UUID
}

func (u *uploadTranslationFileRequest) parseParams() (uploadParams, error) {
	if u == nil {
		return uploadParams{}, errors.New("request is nil")
	}

	var (
		params = uploadParams{data: u.Data, schema: u.Schema}
		err    error
	)

	params.languageTag, err = language.Parse(u.Language)
	if err != nil {
		return uploadParams{}, fmt.Errorf("parse language: %w", err)
	}

	params.serviceID, err = uuid.Parse(u.ServiceId)
	if err != nil {
		return uploadParams{}, fmt.Errorf("parse service uuid: %w", err)
	}

	if u.TranslationFileId == "" {
		return params, nil
	}

	params.translationFileID, err = uuid.Parse(u.TranslationFileId)
	if err != nil {
		return uploadParams{}, fmt.Errorf("parse translate file uuid: %w", err)
	}

	return params, nil
}

// Validates request parameters for UploadTranslationFile.
func (u *uploadParams) validate() error {
	if len(u.data) == 0 {
		return fmt.Errorf("'data' is required")
	}

	// Enforce that schema is present. (Temporal solution)
	if u.schema == translatev1.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *translatev1.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	uploadReq := (*uploadTranslationFileRequest)(req)

	params, err := uploadReq.parseParams()
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

	translationFile := &model.TranslationFile{
		ID:       params.translationFileID,
		Messages: messages,
	}

	switch err := t.repo.SaveTranslationFile(ctx, params.serviceID, translationFile); {
	default:
		return &emptypb.Empty{}, nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Errorf(codes.Internal, err.Error())
	}
}

// ----------------------DownloadTranslationFile-------------------------------

type downloadParams struct {
	languageTag language.Tag
	schema      translatev1.Schema
	serviceID   uuid.UUID
}

func (d *downloadTranslationFileRequest) parseParams() (downloadParams, error) {
	if d == nil {
		return downloadParams{}, errors.New("request is nil")
	}

	var (
		params = downloadParams{schema: d.Schema}
		err    error
	)

	params.serviceID, err = uuid.Parse(d.ServiceId)
	if err != nil {
		return downloadParams{}, fmt.Errorf("parse service uuid: %w", err)
	}

	params.languageTag, err = language.Parse(d.Language)
	if err != nil {
		return downloadParams{}, fmt.Errorf("parse language: %w", err)
	}

	return params, nil
}

func (d *downloadParams) validate() error {
	// Enforce that schema is present. (Temporal solution)
	if d.schema == translatev1.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	return nil
}

func (t *TranslateServiceServer) DownloadTranslationFile(
	ctx context.Context,
	req *translatev1.DownloadTranslationFileRequest,
) (*translatev1.DownloadTranslationFileResponse, error) {
	downloadReq := (*downloadTranslationFileRequest)(req)

	params, err := downloadReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	translationFile, err := t.repo.LoadTranslationFile(ctx, params.serviceID, params.languageTag)

	switch {
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	data, err := MessagesToData(params.schema, translationFile.Messages)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &translatev1.DownloadTranslationFileResponse{Data: data}, nil
}
