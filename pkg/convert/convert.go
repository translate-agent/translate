package convert

import (
	"strings"
)

// convertToMessageFormatSingular wraps the input string with curly braces and returns the modified string.
func convertToMessageFormatSingular(message string) string {
	if message == "" {
		return ""
	}

	message = escapeSpecialChars(message)

	return "{" + message + "}"
}

// escapeSpecialChars escapes special characters in a string.
// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#text
func escapeSpecialChars(message string) string {
	if strings.Contains(message, "{") ||
		strings.Contains(message, "}") ||
		strings.Contains(message, "\\") ||
		strings.Contains(message, "|") {
		r := strings.NewReplacer("{", "\\{", "}", "\\}", "\\", "\\\\", "|", "\\|")
		return r.Replace(message)
	}

	return message
}

// removeEscapeSpecialChars checks if a given string contains specific escape sequences (e.g., "\{"),
// and if found, it replaces them with their unescaped equivalents (e.g., "{" for "\{").
func removeEscapeSpecialChars(message string) string {
	if strings.Contains(message, "\\{") ||
		strings.Contains(message, "\\}") ||
		strings.Contains(message, "\\\\") ||
		strings.Contains(message, "\\|") {
		r := strings.NewReplacer("\\{", "{", "\\}", "}", "\\\\", "\\", "\\|", "|")
		return r.Replace(message)
	}

	return message
}

// removeEnclosingBrackets replaces  '{' prefix and '}' suffix, temporarily maintain only the singular form.
func removeEnclosingBrackets(message string) string {
	message = strings.TrimPrefix(message, "{")
	return strings.TrimSuffix(message, "}")
}
