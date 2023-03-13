package translate

import (
	"context"
	"errors"
	"fmt"

	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	uploadTranslationFileRequest   pb.UploadTranslationFileRequest
	downloadTranslationFileRequest pb.DownloadTranslationFileRequest
)

// ----------------------UploadTranslationFile-------------------------------

type uploadParams struct {
	language language.Tag
	data     []byte
	schema   pb.Schema
}

func (u *uploadTranslationFileRequest) parseParams() (uploadParams, error) {
	if u == nil {
		return uploadParams{}, errors.New("request is nil")
	}

	lang, err := language.Parse(u.Language)
	if err != nil {
		return uploadParams{}, fmt.Errorf("parse language: %w", err)
	}

	return uploadParams{language: lang, data: u.Data, schema: u.Schema}, nil
}

// Validates request parameters for UploadTranslationFile.
func (u *uploadParams) validate() error {
	if len(u.data) == 0 {
		return fmt.Errorf("'data' is required")
	}

	// Enforce that schema is present. (Temporal solution)
	if u.schema == pb.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *pb.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	uploadReq := (*uploadTranslationFileRequest)(req)

	params, err := uploadReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// ----------------------DownloadTranslationFile-------------------------------

type downloadParams struct {
	language language.Tag
	schema   pb.Schema
}

func (u *downloadTranslationFileRequest) parseParams() (downloadParams, error) {
	if u == nil {
		return downloadParams{}, errors.New("request is nil")
	}

	tag, err := language.Parse(u.Language)
	if err != nil {
		return downloadParams{}, fmt.Errorf("parse language: %w", err)
	}

	return downloadParams{language: tag, schema: u.Schema}, nil
}

func (d *downloadParams) validate() error {
	// Enforce that schema is present. (Temporal solution)
	if d.schema == pb.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	return nil
}

func (t *TranslateServiceServer) DownloadTranslationFile(
	ctx context.Context,
	req *pb.DownloadTranslationFileRequest,
) (*pb.DownloadTranslationFileResponse, error) {
	downloadReq := (*downloadTranslationFileRequest)(req)

	params, err := downloadReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.DownloadTranslationFileResponse{}, nil
}
