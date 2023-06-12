package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/language"
)

// Lang returns a random language tag.
func Lang() language.Tag {
	return language.MustParse(gofakeit.LanguageBCP())
}
