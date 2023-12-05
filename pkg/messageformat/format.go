package messageformat

import (
	"regexp"
)

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

/*
GetPlaceholderFormat detects and returns placeholder format for given format specifier,
returns nil if format is not recognized.

Examples:
input: "%d"
output: {

		Re:        regexp.MustCompile(`%(s|d)`),
		NodeExprF: CreateNodeExpr("printf"),
	}
*/
func GetPlaceholderFormat(formatSpecifier string) *PlaceholderFormat {
	for _, v := range placeholderFormats {
		if v.Re.MatchString(formatSpecifier) {
			return &v
		}
	}

	return nil
}
