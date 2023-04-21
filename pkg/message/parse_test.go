package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedErr error
		expected    []interface{}
	}{
		{
			name:     "text node",
			input:    "{Hello, world!}",
			expected: []interface{}{NodeText{Text: "Hello, world!"}},
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
			input: "match {$count :number} when 1 {Buy one apple!} when * {Buy {$count} apples!} ",
			expected: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Buy one apple!"}}},
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
			name:        "missing opening for text node",
			input:       "Hello, world!}",
			expectedErr: fmt.Errorf("unknown token: %+v", Token{Type: TokenTypeText, Value: "Hello, world!}"}),
		},
		{
			name:        "missing closing for text node",
			input:       "{Hello, world!",
			expectedErr: fmt.Errorf("text does not end with '}'"),
		},
		{
			name:        "function does not start with variable",
			input:       "match {:number} when * {Hello, world!}",
			expectedErr: fmt.Errorf("function does not start with variable"),
		},
		{
			name:        "variable starts with space",
			input:       "match {$ count} when * {Hello, world!}",
			expectedErr: fmt.Errorf("variable does not start with '$'"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := Parse(tt.input)

			if tt.expectedErr != nil {
				assert.Errorf(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}
