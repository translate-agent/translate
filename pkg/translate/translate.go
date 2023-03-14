package translate

import (
	pb "go.expect.digital/translate/pkg/pb/translate/v1"
)

type TranslateServiceServer struct {
	pb.UnimplementedTranslateServiceServer
}
