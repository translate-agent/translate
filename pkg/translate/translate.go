package translate

import (
	pb "go.expect.digital/translate/pkg/server/translate/v1"
)

type TranslateServiceServer struct {
	pb.UnimplementedTranslateServiceServer
}
