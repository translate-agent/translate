package convert

import (
	"strings"

	"go.expect.digital/translate/pkg/model"
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

// removeEnclosingBrackets replaces '{' and '}', temporarily maintain only the singular form.
func removeEnclosingBrackets(message string) string {
	replacer := strings.NewReplacer("{", "", "}", "")
	return replacer.Replace(message)
}

/*
getStatus returns the status of a message based on the original and fuzzy flags and the message string.

  - Scenario 1: is original -> translated
  - Scenario 2: not original and msg is not empty -> translated
  - Scenario 3: not original and msg is empty -> untranslated
  - Scenario 4: is fuzzy -> fuzzy
*/
func getStatus(msg string, original, fuzzy bool) (status model.MessageStatus) {
	switch {
	default:
		status = model.MessageStatusUntranslated
	case fuzzy:
		status = model.MessageStatusFuzzy
	case original, !original && (msg != "" && msg != "{}"):
		status = model.MessageStatusTranslated
	}

	return
}
