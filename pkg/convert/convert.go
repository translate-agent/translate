package convert

import ast "go.expect.digital/mf2/parse"

func patternsToSimpleMsg(patterns []ast.PatternPart) string {
	var text string

	for _, p := range patterns {
		if textPattern, ok := p.(ast.Text); ok {
			text += string(textPattern)
		}
	}

	return text
}
