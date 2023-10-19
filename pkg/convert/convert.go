package convert

import (
	"errors"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/messageformat"
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

func getMsg(message string) (string, error) {
	nodes, err := messageformat.Parse(message)
	if err != nil {
		return "", fmt.Errorf("parse message: %w", err)
	}

	for _, node := range nodes {
		nodeTxt, ok := node.(messageformat.NodeText)
		if !ok {
			return "", errors.New("convert node to messageformat.NodeText")
		} else {
			return nodeTxt.Text, nil
		}
	}

	return "", errors.New("input message is empty")
}
