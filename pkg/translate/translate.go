package translate

import tpb "go.expect.digital/translate/pkg/pb/translate/v1"

type TranslateServiceServer struct {
	tpb.UnimplementedTranslateServiceServer
}
