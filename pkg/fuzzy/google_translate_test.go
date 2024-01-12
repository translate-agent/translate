package fuzzy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SplitTextByPlaceholder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "{$0}",
			expected: []string{"{$0}"},
		},
		{
			input:    " {$0}",
			expected: []string{" ", "{$0}"},
		},
		{
			input:    "{$0} ",
			expected: []string{"{$0}", " "},
		},
		{
			input:    "{$0}{$1}{$2}",
			expected: []string{"{$0}", "{$1}", "{$2}"},
		},
		{
			input:    "{$0}{$0}",
			expected: []string{"{$0}", "{$0}"},
		},
		{
			input:    "Hello {$0} {$1}!",
			expected: []string{"Hello ", "{$0}", " ", "{$1}", "!"},
		},
		{
			input:    "Hello {$0} {$1}! Welcome to {$2}.",
			expected: []string{"Hello ", "{$0}", " ", "{$1}", "! Welcome to ", "{$2}", "."},
		},
		{
			input:    "Hello {$0} {$1}! Welcome to {$2}.",
			expected: []string{"Hello ", "{$0}", " ", "{$1}", "! Welcome to ", "{$2}", "."},
		},
	}

	for i, test := range tests {
		test := test

		t.Run(fmt.Sprintf("test number %d", i), func(t *testing.T) {
			t.Parallel()

			actual := splitTextByPlaceholder(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
