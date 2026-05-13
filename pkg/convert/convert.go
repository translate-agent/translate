package convert

import "go.expect.digital/mf2/parse"

func patternsToSimpleMsg(patterns []parse.PatternPart) string {
	var text string

	for _, p := range patterns {
		if textPattern, ok := p.(parse.Text); ok {
			text += string(textPattern)
		}
	}

	return text
}
