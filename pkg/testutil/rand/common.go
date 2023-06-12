package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/language"
)

// Lang returns a random language tag.
func Lang() language.Tag {
	return language.MustParse(gofakeit.LanguageBCP())
}

// Langs returns a slice of n random unique language tags.
func Langs(n uint) []language.Tag {
	languagesUsed := make(map[language.Tag]bool, n)
	tags := make([]language.Tag, 0, n)

	for uint(len(tags)) < n {
		lang := Lang()
		if !languagesUsed[lang] {
			languagesUsed[lang] = true

			tags = append(tags, lang)
		}
	}

	return tags
}
