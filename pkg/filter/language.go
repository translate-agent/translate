package filter

import "golang.org/x/text/language"

// LanguageTags removes duplicates and empty tags returns slice of unique language tags.
func LanguageTags(lt []language.Tag) []language.Tag {
	uniques := make([]language.Tag, 0, len(lt))

	m := make(map[language.Tag]struct{}, len(lt))

	for _, v := range lt {
		if _, ok := m[v]; ok || v == language.Und {
			continue
		}

		m[v] = struct{}{}

		uniques = append(uniques, v)
	}

	return uniques
}
