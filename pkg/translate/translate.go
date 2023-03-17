package translate

import translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"

type TranslateServiceServer struct {
	translatev1.UnimplementedTranslateServiceServer
}
