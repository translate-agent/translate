package translate

import (
	"context"

	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TranslateServiceServer struct {
	pb.UnimplementedTranslateServiceServer
}

func (t *TranslateServiceServer) UploadTranslationFile(
	ctx context.Context,
	req *pb.UploadTranslationFileRequest,
) (*pb.UploadTranslationFileResponse, error) {
	var (
		reqLanguage = req.GetLanguage()
		reqData     = req.GetData()
		reqSchema   = req.GetSchema()
	)

	if len(reqLanguage) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "'language' is missing")
	}

	languageTag, err := language.Parse(reqLanguage)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse language: %s", err)
	}

	_ = languageTag
	_ = reqSchema
	_ = reqData

	// convert from `schema` to our messages

	return &pb.UploadTranslationFileResponse{
		Messages: []*pb.Messages{},
	}, nil
}
