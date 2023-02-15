package translate

import (
	"context"
	"fmt"

	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TranslateServiceServer struct {
	pb.UnimplementedTranslateServiceServer
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *pb.UploadTranslationFileRequest,
) (*emptypb.Empty, error) {
	var (
		reqLanguage = req.GetLanguage()
		reqData     = req.GetData()
		reqSchema   = req.GetSchema()
	)

	if len(reqLanguage) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "'language' is required")
	}

	if len(reqData) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "'data' is required")
	}

	languageTag, err := language.Parse(reqLanguage)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse language: %s", err)
	}

	_ = languageTag
	_ = reqSchema
	_ = reqData

	// convert from `schema` to our messages

	return &emptypb.Empty{}, nil
}

// ----------------------DownloadTranslationFile-------------------------------

type LanguageData struct {
	Tag language.Tag
	Str string
}

type DownloadParams struct {
	Language LanguageData
	Schema   pb.Schema
}

// Validates parameters for DownloadTranslationFile.
func (d *DownloadParams) validate() error {
	if len(d.Language.Str) == 0 {
		return fmt.Errorf("'language' is required")
	}

	var err error

	if d.Language.Tag, err = language.Parse(d.Language.Str); err != nil {
		return fmt.Errorf("parsing language '%s': %w", d.Language.Str, err)
	}

	return nil
}

func (t *TranslateServiceServer) DownloadTranslationFile(
	ctx context.Context,
	req *pb.DownloadTranslationFileRequest,
) (*pb.DownloadTranslationFileResponse, error) {
	reqParams := DownloadParams{
		Schema: req.GetSchema(),
		Language: LanguageData{
			Str: req.GetLanguage(),
		},
	}

	if err := reqParams.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.DownloadTranslationFileResponse{}, nil
}
