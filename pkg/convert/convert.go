package convert

import "strings"

// convertToMessageFormatSingular wraps the input string with curly braces and returns the modified string.
func convertToMessageFormatSingular(message string) string {
	if message == "" {
		return ""
	}

	if strings.Contains(message, "{") ||
		strings.Contains(message, "}") ||
		strings.Contains(message, "|") ||
		strings.Contains(message, "\\") {
		r := strings.NewReplacer("{", "\\{", "}", "\\}", "|", "\\|", "\\", "\\\\")
		message = r.Replace(message)
	}

	return "{" + message + "}"
}

// removeEnclosingBrackets replaces '{' and '}', temporarily maintain only the singular form.
func removeEnclosingBrackets(message string) string {
	replacer := strings.NewReplacer("{", "", "}", "")
	return replacer.Replace(message)
}
