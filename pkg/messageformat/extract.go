package messageformat

import (
	"regexp"
	"strings"
)

// specialChars is a list of characters that need to be escaped in the node.Text,
// because they are either syntax characters or reserved keywords.
var specialChars = map[rune]struct{}{
	// Syntax characters
	'{':  {},
	'}':  {},
	'\\': {},
	'|':  {},
	// Reserved keywords (Future syntax characters)
	'!': {},
	'@': {},
	'#': {},
	'%': {},
	'*': {},
	'<': {},
	'>': {},
	'/': {},
	'?': {},
	'~': {},
}

// EscapeSpecialChars escapes special characters in the given text.
func EscapeSpecialChars(text string) string {
	var sb strings.Builder

	for _, c := range text {
		if _, ok := specialChars[c]; ok {
			sb.WriteRune('\\')
		}

		sb.WriteRune(c)
	}

	return sb.String()
}

type PlaceholderFormat struct {
	Re        *regexp.Regexp
	NodeExprF func(string, []int) NodeExpr
}

// printfVerbs maps printf verbs to their corresponding type.
var printfVerbs = map[string]string{
	"s": "string",
	"d": "int",
	"f": "float",
}

// placeholderFormats maps placeholder formats to their corresponding regex and NodeExpr factory function.
var placeholderFormats = map[string]PlaceholderFormat{
	"pythonVar": { // hello %(var)s | hello %(var)d
		Re:        regexp.MustCompile(`%\((\w+)\)([sd])`),
		NodeExprF: CreateNodeExpr("pythonVar"),
	},
	"printf": { // hello %s | hello %d
		Re:        regexp.MustCompile(`%(s|d)`),
		NodeExprF: CreateNodeExpr("printf"),
	},
	"bracketVar": { // hello {var} | hello {0}
		Re:        regexp.MustCompile(`\{(\w+)\}`),
		NodeExprF: CreateNodeExpr("bracketVar"),
	},
	"emptyBracket": { // hello {}
		Re:        regexp.MustCompile(`\{\}`),
		NodeExprF: CreateNodeExpr("emptyBracket"),
	},
}

func GetPlaceholderFormat(s string) *PlaceholderFormat {
	for _, v := range placeholderFormats {
		if v.Re.MatchString(s) {
			return &v
		}
	}

	return nil
}

// CreateNodeExpr creates a NodeExpr factory function for the given format.
//
// Example:
//
//	CreateNodeExpr("pythonVar")("hello %(var)s", re.FindStringSubmatchIndex("hello %(var)s", -1))
//
// Result:
//
//	NodeExpr{
//		Function: NodeFunction{
//				Name: "Placeholder",
//				Options: []NodeOption{
//					{Name: "format", Value: "pythonVar"},
//					{Name: "name", Value: "var"},
//					{Name: "type", Value: "string"},
//					},
//				},
//			}
func CreateNodeExpr(format string, opts ...NodeOption) func(string, []int) NodeExpr {
	return func(text string, indices []int) NodeExpr {
		var varName, varType string

		switch format {
		case "pythonVar":
			varName, varType = text[indices[2]:indices[3]], printfVerbs[text[indices[4]:indices[5]]]
		case "printf":
			varType = printfVerbs[text[indices[2]:indices[3]]]
		case "bracketVar":
			varName = text[indices[2]:indices[3]]
		case "emptyBracket":
		}

		var options []NodeOption

		// Add placeholder format
		options = append(options, NodeOption{Name: "format", Value: format})

		// Add variable name
		if varName != "" {
			options = append(options, NodeOption{Name: "name", Value: varName})
		}

		// Add variable type
		if varType != "" {
			options = append(options, NodeOption{Name: "type", Value: varType})
		}

		options = append(options, opts...)

		return NodeExpr{Function: NodeFunction{Name: "Placeholder", Options: options}}
	}
}
