package convert

import "strings"

// convertToMessageFormatSingular wraps the input string with curly braces and returns the modified string.
func convertToMessageFormatSingular(message string) string {
	if message == "" {
		return ""
	}

	message = escapeSpecialChars(message)
	message = escapePipe(message)

	return "{" + message + "}"
}

// escapeSpecialChars escapes special characters in a string.
func escapeSpecialChars(message string) string {
	if strings.Contains(message, "{") ||
		strings.Contains(message, "}") ||
		strings.Contains(message, "\\") {
		r := strings.NewReplacer("{", "\\{", "}", "\\}", "\\", "\\\\")
		message = r.Replace(message)
	}

	return message
}

// removeEnclosingBrackets replaces '{' and '}', temporarily maintain only the singular form.
func removeEnclosingBrackets(message string) string {
	replacer := strings.NewReplacer("{", "", "}", "")
	return replacer.Replace(message)
}

// escapePipe escapes single pipe characters '|' in a given string
// but leaves double pipe characters '||' untouched.
func escapePipe(input string) string {
	var output strings.Builder

	for i := 0; i < len(input); i++ {
		if input[i] == '|' {
			// Check if the next character is also a '|'
			if i < len(input)-1 && input[i+1] == '|' {
				output.WriteByte('|')
				output.WriteByte('|')
				i++ // Skip the next '|'
			} else {
				output.WriteString("\\|")
			}
		} else {
			output.WriteByte(input[i])
		}
	}

	return output.String()
}
