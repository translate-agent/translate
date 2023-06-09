package convert

import (
	"fmt"
	"strings"
)

func convertToMessageFormatPlural(plurals []string) string {
	var sb strings.Builder

	sb.WriteString("match {$count :number}\n")

	for i, plural := range plurals {
		line := strings.ReplaceAll(strings.TrimSpace(plural), "%d", "{$count}")

		var count string

		if i == len(plurals)-1 {
			count = "*"
		} else {
			count = fmt.Sprintf("%d", i+1)
		}

		sb.WriteString(fmt.Sprintf("when %s {%s}\n", count, line))
	}

	return sb.String()
}

func convertToMessageFormatSingular(message string) string {
	if message == "" {
		return ""
	}

	return "{" + message + "}"
}
