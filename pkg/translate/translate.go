package translate

import (
	"context"
	"fmt"

	"go.expect.digital/translate/pkg/repo"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TranslateServiceServer struct {
	pb.UnimplementedTranslateServiceServer
	repo repo.Repo
}

func NewTranslateServiceServer(r repo.Repo) *TranslateServiceServer {
	return &TranslateServiceServer{repo: r}
}

// ----------------------UploadTranslationFile-------------------------------

type uploadParams struct {
	Language language.Tag
	Data     []byte
	Schema   pb.Schema
}

func parseUploadParams(req *pb.UploadTranslationFileRequest) (*uploadParams, error) {
	reqLanguage := req.GetLanguage()

	tag, err := language.Parse(reqLanguage)
	if err != nil {
		return nil, fmt.Errorf("parse language '%s': %w", reqLanguage, err)
	}

	params := uploadParams{
		Language: tag,
		Data:     req.GetData(),
		Schema:   req.GetSchema(),
	}

	return &params, nil
}

// Validates request parameters for UploadTranslationFile.
func (u *uploadParams) validate() error {
	if len(u.Data) == 0 {
		return fmt.Errorf("'data' is required")
	}

	// Enforce that schema is present. (Temporal solution)
	if u.Schema == pb.Schema_UNSPECIFIED {
		return fmt.Errorf("'schema' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *pb.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	params, err := parseUploadParams(req)
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
	Language language.Tag
	Schema   pb.Schema
}

func parseDownloadParams(req *pb.DownloadTranslationFileRequest) (*downloadParams, error) {
	reqLanguage := req.GetLanguage()

	tag, err := language.Parse(reqLanguage)
	if err != nil {
		return nil, fmt.Errorf("parse language '%s': %w", reqLanguage, err)
	}

	params := downloadParams{
		Language: tag,
		Schema:   req.GetSchema(),
	}

	return &params, nil
}

func (t *TranslateServiceServer) DownloadTranslationFile(
	ctx context.Context,
	req *pb.DownloadTranslationFileRequest,
) (*pb.DownloadTranslationFileResponse, error) {
	params, err := parseDownloadParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	_ = params

	return &pb.DownloadTranslationFileResponse{}, nil
}
