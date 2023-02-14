package translate

import (
	"context"

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
