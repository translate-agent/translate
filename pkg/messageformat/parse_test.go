package messageformat

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, input string
		expectedErr error
		expected    []interface{}
	}{
		{
			name:     "empty",
			input:    "",
			expected: []interface{}(nil),
		},
		{
			name:     "empty expr",
			input:    "{}",
			expected: []interface{}(nil),
		},
		{
			name:     "expr with text",
			input:    "{Hello, World!}",
			expected: []interface{}{NodeText{Text: "Hello, World!"}},
		},
		{
			name:  "match",
			input: "match {$count} when * {Hello, world!}",
			expected: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world!"}}},
					},
				},
			},
		},
		{
			name:  "match with function",
			input: "match {$count :number} when * {Hello, world!}",
			expected: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world!"}}},
					},
				},
			},
		},
		{
			name:  "match with multiple variants",
			input: "match {$count :number} when 1 {Hello, friend!} when * {Hello, friends!} ",
			expected: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Hello, friend!"}}},
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, friends!"}}},
					},
				},
			},
		},
		{
			name:  "match with plurals",
			input: "match {$count :number} when 1 {Buy one \\\\ apple!} when * {Buy {$count} apples!} ",
			expected: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Buy one \\ apple!"}}},
						{Keys: []string{"*"}, Message: []interface{}{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeText{Text: " apples!"},
						}},
					},
				},
			},
		},

		{
			name: "match with two variables in variant",
			input: "match {$count :number} " +
				"when 0 {No apples!} " +
				"when 1 {Buy {$count}{$counts} apple!} " +
				"when * {Buy {$count} apples 2!} ",
			expected: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"0"}, Message: []interface{}{NodeText{Text: "No apples!"}}},
						{Keys: []string{"1"}, Message: []interface{}{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeVariable{Name: "counts"},
							NodeText{Text: " apple!"},
						}},
						{Keys: []string{"*"}, Message: []interface{}{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeText{Text: " apples 2!"},
						}},
					},
				},
			},
		},
		{
			name:        "invalid expr",
			input:       "match $count :number} ",
			expectedErr: fmt.Errorf("expression does not start with \"{\""),
		},
		{
			name:     "input with curly braces in it",
			input:    `{Chart [\{\}] was added to dashboard [\{\}]}`,
			expected: []interface{}{NodeText{Text: "Chart [{}] was added to dashboard [{}]"}},
		},
		{
			name:     "input with plus sign in it ",
			input:    `{+ vēl %s}`,
			expected: []interface{}{NodeText{Text: "+ vēl %s"}},
		},
		{
			name:     "input with minus sign in it ",
			input:    `{- vēl %s}`,
			expected: []interface{}{NodeText{Text: "- vēl %s"}},
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, err := Parse(test.input)

			if test.expectedErr != nil {
				require.Errorf(t, err, test.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, l)
		})
	}
}
