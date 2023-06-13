package convert

import (
	"strings"
)

// convertToMessageFormatSingular wraps the input string with curly braces and returns the modified string.
func convertToMessageFormatSingular(message string) string {
	if message == "" {
		return ""
	}

	return "{" + message + "}"
}

// convertFromMessageFormatStrToStr function replaces '{' and '}' only if message starts with '{' and ends with '}'.
func convertFromMessageFormatStrToStr(message string) string {
	if strings.HasPrefix(message, "{") && strings.HasSuffix(message, "}") {
		str := strings.TrimSuffix(message, "}")
		str = strings.TrimPrefix(str, "{")

		return str
	}

	return message
}
