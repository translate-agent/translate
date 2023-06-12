package rand

import (
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/language"
)

// Lang returns a random language tag.
func Lang() language.Tag {
	return language.MustParse(gofakeit.LanguageBCP())
}

// Langs returns a slice of random unique language tags.
func Langs(count uint) []language.Tag {
	languagesUsed := make(map[language.Tag]bool, count)

	tags := make([]language.Tag, 0, count)

	for i := uint(0); i < count; i++ {
		lang := Lang()

		for languagesUsed[lang] {
			lang = Lang()
		}

		languagesUsed[lang] = true

		tags = append(tags, lang)
	}

	return tags
}
