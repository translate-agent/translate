package convert

// convertToMessageFormatSingular wraps the input string with curly braces and returns the modified string.
func convertToMessageFormatSingular(message string) string {
	if message == "" {
		return ""
	}

	return "{" + message + "}"
}
