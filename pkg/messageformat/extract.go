package messageformat

import (
	"fmt"
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

const maxOptionCount = 3 // maximum size of []Node.Option, when parsing a placeholder

type placeholderFormat struct {
	re        *regexp.Regexp
	nodeExprF func(string, []int) NodeExpr
}

// printfVerbs maps printf verbs to their corresponding type.
var printfVerbs = map[string]string{
	"s": "string",
	"d": "int",
	"f": "float",
}

// placeholderFormats maps placeholder formats to their corresponding regex and NodeExpr factory function.
var placeholderFormats = map[string]placeholderFormat{
	"pythonVar": { // hello %(var)s | hello %(var)d
		re:        regexp.MustCompile(`%\((\w+)\)([sd])`),
		nodeExprF: createNodeExpr("pythonVar"),
	},
	"printf": { // hello %s | hello %d
		re:        regexp.MustCompile(`%(s|d)`),
		nodeExprF: createNodeExpr("printf"),
	},
	"bracketVar": { // hello {var} | hello {0}
		re:        regexp.MustCompile(`\{(\w+)\}`),
		nodeExprF: createNodeExpr("bracketVar"),
	},
	"emptyBracket": { // hello {}
		re:        regexp.MustCompile(`\{\}`),
		nodeExprF: createNodeExpr("emptyBracket"),
	},
}

// messageToAST converts the given message to an MF2 AST.
func (pf *placeholderFormat) messageToAST(message string) AST {
	// If pf is nil, all message is a NodeText
	if pf.re == nil {
		return AST{NodeText{Text: escapeSpecialChars(message)}}
	}

	// If pf is not nil then message contains placeholders, so we need to parse it
	var (
		currentIdx int
		ast        AST // Only NodeText and NodeExpr
	)

	for _, matchIndices := range pf.re.FindAllStringSubmatchIndex(message, -1) {
		// If there is text before the current match, add it as a NodeText
		if matchIndices[0] != currentIdx {
			text := message[currentIdx:matchIndices[0]]
			ast.Append(NodeText{Text: escapeSpecialChars(text)})
		}

		// Convert the current match to a NodeExpr and add it to the nodes
		ast.Append(pf.nodeExprF(message, matchIndices))
		currentIdx = matchIndices[1]
	}

	// If there is text after the last match, add it as a NodeText
	if currentIdx < len(message) {
		text := message[currentIdx:]
		ast.Append(NodeText{Text: escapeSpecialChars(text)})
	}

	return ast
}

// createNodeExpr creates a NodeExpr factory function for the given format.
//
// Example:
//
//	createNodeExpr("pythonVar")("hello %(var)s", re.FindStringSubmatchIndex("hello %(var)s", -1))
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
func createNodeExpr(format string) func(string, []int) NodeExpr {
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

		options := make([]NodeOption, 0, maxOptionCount)

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

		return NodeExpr{Function: NodeFunction{Name: "Placeholder", Options: options}}
	}
}

// escapeSpecialChars escapes special characters in the given text.
func escapeSpecialChars(text string) string {
	var sb strings.Builder

	for _, c := range text {
		if _, ok := specialChars[c]; ok {
			sb.WriteRune('\\')
		}

		sb.WriteRune(c)
	}

	return sb.String()
}

// ToMessageFormat2 converts the given string to a MessageFormat2 compliant string.
//
// Examples:
//
//	"Hello world" -> "{Hello world}" // normal text
//	"Hello world!" -> "{Hello world\\!}" // text with special character
//	"Hello {user}" -> "{Hello {:Placeholder format=bracketVar name=user} world}" // text with bracket placeholder
//	"Hello %s". -> "{Hello {:Placeholder format=printf type=string}.}" // text with printf placeholder
func ToMessageFormat2(message string) (string, error) {
	if message == "" {
		return "", nil
	}

	var pf placeholderFormat

	for _, v := range placeholderFormats {
		// Find the first placeholder format that matches the message
		if v.re.MatchString(message) {
			pf = v
			break
		}
	}

	data, err := pf.messageToAST(message).MarshalText()
	if err != nil {
		return "", fmt.Errorf("encode message format v2: %w", err)
	}

	return string(data), err
}
