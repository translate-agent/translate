package translate

import (
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
)

type TranslateServiceServer struct {
	translatev1.UnimplementedTranslateServiceServer
	repo repo.Repo
}

func NewTranslateServiceServer(r repo.Repo) *TranslateServiceServer {
	return &TranslateServiceServer{repo: r}
}
