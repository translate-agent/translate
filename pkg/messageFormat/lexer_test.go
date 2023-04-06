package messageFormat

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLexMessage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		ex    []Token
	}{
		{
			name:  "simple",
			input: "Hello, {name}!",
			ex: []Token{
				{Type: literal, Value: "Hello,"},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "name"},
				{Type: placeholderClose, Value: "}"},
				{Type: literal, Value: "!"},
			},
		},

		{
			name:  "inner placeholders",
			input: "{count, plural, one{1 message} other{# messages :ident}}",
			ex: []Token{
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "count,"},
				{Type: text, Value: "plural,"},
				{Type: text, Value: "one"},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "1"},
				{Type: text, Value: "message"},
				{Type: placeholderClose, Value: "}"},
				{Type: text, Value: "other"},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "#"},
				{Type: text, Value: "messages"},
				{Type: variable, Value: ":ident"},
				{Type: placeholderClose, Value: "}"},
				{Type: placeholderClose, Value: "}"},
			},
		},
		{
			name:  "plurals with keyword",
			input: "match {$count :number +add -sub} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			ex: []Token{
				{Type: keyword, Value: "match"},
				{Type: placeholderOpen, Value: "{"},
				{Type: variable, Value: "$count"},
				{Type: variable, Value: ":number"},
				{Type: variable, Value: "+add"},
				{Type: variable, Value: "-sub"},
				{Type: placeholderClose, Value: "}"},
				{Type: keyword, Value: "when"},
				{Type: literal, Value: "1"},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "You"},
				{Type: text, Value: "have"},
				{Type: text, Value: "one"},
				{Type: text, Value: "notification."},
				{Type: placeholderClose, Value: "}"},
				{Type: keyword, Value: "when"},
				{Type: literal, Value: "*"},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "You"},
				{Type: text, Value: "have"},
				{Type: placeholderOpen, Value: "{"},
				{Type: variable, Value: "$count"},
				{Type: placeholderClose, Value: "}"},
				{Type: text, Value: "notifications."},
				{Type: placeholderClose, Value: "}"},
			},
		},
		{
			name:  "plurals",
			input: "match {$count :number} when 1 {You have one notification.} when * {You have {$count} notifications.}",
			ex: []Token{
				{Type: keyword, Value: "match "},
				{Type: placeholderOpen, Value: "{"},
				{Type: variable, Value: "count "},
				{Type: function, Value: "number"},
				{Type: placeholderClose, Value: "}"},
				{Type: keyword, Value: " when "},
				{Type: literal, Value: "1 "},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "You have one notification."},
				{Type: placeholderClose, Value: "}"},
				{Type: keyword, Value: " when "},
				{Type: literal, Value: "* "},
				{Type: placeholderOpen, Value: "{"},
				{Type: text, Value: "You have "},
				{Type: placeholderOpen, Value: "{"},
				{Type: variable, Value: "count"},
				{Type: placeholderClose, Value: "}"},
				{Type: text, Value: " notifications."},
				{Type: placeholderClose, Value: "}"},
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tokenize(tt.input)
			fmt.Printf("res: %#v\n", result)
			fmt.Printf("exp: %#v\n", tt.ex)
			assert.Equal(t, tt.ex, result)
		})
	}
}
