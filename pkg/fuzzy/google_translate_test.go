package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SplitTextByPlaceholder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "",
			want:  []string(nil),
		},
		{
			input: "a",
			want:  []string{"a"},
		},
		{
			input: "Welcome to",
			want:  []string{"Welcome to"},
		},
		{
			input: "{$0}",
			want:  []string{"{$0}"},
		},
		{
			input: " {$0}",
			want:  []string{" ", "{$0}"},
		},
		{
			input: "{$0} ",
			want:  []string{"{$0}", " "},
		},
		{
			input: "{$0}{$1}{$2}",
			want:  []string{"{$0}", "{$1}", "{$2}"},
		},
		{
			input: "{$0}{$0}",
			want:  []string{"{$0}", "{$0}"},
		},
		{
			input: "Hello {$0} {$1}!",
			want:  []string{"Hello ", "{$0}", " ", "{$1}", "!"},
		},
		{
			input: "Hello {$0} {$1}! Welcome to {$2}.",
			want:  []string{"Hello ", "{$0}", " ", "{$1}", "! Welcome to ", "{$2}", "."},
		},
		{
			input: "Hello {$0} {$1}! Welcome to {$2}.",
			want:  []string{"Hello ", "{$0}", " ", "{$1}", "! Welcome to ", "{$2}", "."},
		},
		{
			input: "{$100} {$200}! Welcome to {$300}",
			want:  []string{"{$100}", " ", "{$200}", "! Welcome to ", "{$300}"},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			got := splitTextByPlaceholder(test.input)
			assert.Equal(t, test.want, got)
		})
	}
}
