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

	return "{{" + message + "}}"
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

func getMsg(message string) (string, error) {
	nodes, err := messageformat.Parse(message)
	if err != nil {
		return "", fmt.Errorf("parse message: %w", err)
	}

	for _, node := range nodes {
		nodeTxt, ok := node.(messageformat.NodeText)
		if ok {
			return nodeTxt.Text, nil
		}
	}

	return "", errors.New("input message is empty")
}

func ptr[T any](v T) *T {
	return &v
}
