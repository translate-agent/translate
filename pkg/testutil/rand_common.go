package testutil

import (
	"github.com/brianvoe/gofakeit/v6"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

// RandLang returns a random language tag.
func RandLang() language.Tag {
	return language.MustParse(gofakeit.LanguageBCP())
}

// RandSchema returns a random traslatev1.Schema.
func RandSchema() translatev1.Schema {
	return translatev1.Schema(gofakeit.IntRange(1, len(translatev1.Schema_name)))
}
