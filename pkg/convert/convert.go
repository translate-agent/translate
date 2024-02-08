package convert

import ast "go.expect.digital/mf2/parse"

func patternsToMsg(patterns []ast.Pattern) string {
	var text string

	for _, p := range patterns {
		if textPattern, ok := p.(ast.TextPattern); ok {
			text += string(textPattern)
		}
	}

	return text
}
