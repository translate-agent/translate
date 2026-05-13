package convert

import (
	"strings"

	"go.expect.digital/mf2/parse"
)

func patternsToSimpleMsg(patterns []parse.PatternPart) string {
	var sb strings.Builder

	for _, p := range patterns {
		if textPattern, ok := p.(parse.Text); ok {
			sb.WriteString(string(textPattern))
		}
	}

	return sb.String()
}
