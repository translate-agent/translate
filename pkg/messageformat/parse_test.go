package messageformat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, input string
		expected    []interface{}
	}{
		{
			name:     "empty",
			input:    "",
			expected: []interface{}{},
		},
		{
			name:     "empty",
			input:    "{}",
			expected: []interface{}{},
		},
		{
			name:     "expr with text",
			input:    "{Hello, World!}",
			expected: []interface{}{},
		},
		{
			name:     "expr with variable",
			input:    "{$count}",
			expected: []interface{}{NodeVariable{Name: "count"}},
		},
		{
			name:     "expr with function",
			input:    "{:rand}",
			expected: []interface{}{NodeFunction{Name: "rand"}},
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, _ := Parse(test.input)

			assert.Equal(t, test.expected, l)
		})
	}
}
