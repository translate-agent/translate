package server

import (
	"go.expect.digital/translate/pkg/fuzzy"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
)

type TranslateServiceServer struct {
	translatev1.UnimplementedTranslateServiceServer
	repo       repo.Repo
	translator fuzzy.Translator
}

func NewTranslateServiceServer(r repo.Repo, translator fuzzy.Translator) *TranslateServiceServer {
	return &TranslateServiceServer{repo: r, translator: translator}
}
