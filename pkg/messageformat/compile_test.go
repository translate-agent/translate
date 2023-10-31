package messageformat

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Compile(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		err error

		name     string
		expected string
		input    AST
	}{
		{
			name:     "no nodes",
			input:    []interface{}{},
			expected: "",
			err:      errors.New("AST must contain at least one node"),
		},
		{
			name:     "single text node",
			input:    []interface{}{NodeText{Text: "Hello, World\\!"}},
			expected: "{Hello, World\\!}",
		},
		{
			name: "multiple text nodes",
			input: []interface{}{
				NodeText{Text: "Hello,"},
				NodeText{Text: " "},
				NodeText{Text: "World"},
				NodeText{Text: "\\!"},
			},
			expected: "{Hello, World\\!}",
		},
		{
			name:     "text with special characters",
			input:    []interface{}{NodeText{Text: "\\{\\}\\|\\!\\@\\#\\%\\*\\<\\>\\/\\?\\~\\\\"}},
			expected: "{\\{\\}\\|\\!\\@\\#\\%\\*\\<\\>\\/\\?\\~\\\\}",
		},
		{
			name:     "text contains plus sign",
			input:    []interface{}{NodeText{Text: "+ vēl \\%s"}},
			expected: "{+ vēl \\%s}",
		},
		{
			name:     "text contains minus sign",
			input:    []interface{}{NodeText{Text: "- vēl \\%s"}},
			expected: "{- vēl \\%s}",
		},
		{
			name: "match, single variant",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world\\!"}}},
					},
				},
			},
			expected: "match {$count} when * {Hello, world\\!}",
		},
		{
			name: "match, single variant with function",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world\\!"}}},
					},
				},
			},
			expected: "match {$count :number} when * {Hello, world\\!}",
		},
		{
			name: "match, multiple variants",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Hello, friend\\!"}}},
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, friends\\!"}}},
					},
				},
			},
			expected: "match {$count :number} when 1 {Hello, friend\\!} when * {Hello, friends\\!}",
		},
		{
			name: "match, multiple variants 2",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Buy one \\\\ apple!"}}},
						{Keys: []string{"*"}, Message: []interface{}{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeText{Text: " apples\\!"},
						}},
					},
				},
			},
			expected: "match {$count :number} when 1 {Buy one \\\\ apple!} when * {Buy {$count} apples\\!}",
		},
		{
			name: "match, variant with two variables",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"0"}, Message: []interface{}{NodeText{Text: "No apples\\!"}}},
						{Keys: []string{"1"}, Message: []interface{}{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeVariable{Name: "counts"},
							NodeText{Text: " apple\\!"},
						}},
						{Keys: []string{"*"}, Message: []interface{}{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeText{Text: " apples 2\\!"},
						}},
					},
				},
			},
			expected: "match {$count :number} " +
				"when 0 {No apples\\!} " +
				"when 1 {Buy {$count}{$counts} apple\\!} " +
				"when * {Buy {$count} apples 2\\!}",
		},
		{
			name: "match, three variants",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "userName"}, Function: NodeFunction{Name: "hasCase"}}},
					Variants: []NodeVariant{
						{
							Keys: []string{"0", "vocative"},
							Message: []interface{}{
								NodeText{Text: "Hello, "},
								NodeExpr{
									NodeVariable{Name: "userName"},
									NodeFunction{
										Name: "person",
										Options: []NodeOption{
											{Name: "case", Value: "vocative"},
											{Name: "format", Value: "printf"},
											{Name: "type", Value: "string"},
										},
									},
								},
								NodeText{Text: "\\!"},
							},
						},
						{
							Keys: []string{"1", "accusative"},
							Message: []interface{}{
								NodeText{Text: "Please welcome "},
								NodeExpr{
									NodeVariable{Name: "userName"},
									NodeFunction{
										Name: "person",
										Options: []NodeOption{
											{Name: "case", Value: "accusative"},
											{Name: "format", Value: "printf"},
											{Name: "type", Value: "string"},
										},
									},
								},
								NodeText{Text: "\\!"},
							},
						},
						{
							Keys: []string{"*", "neutral"},
							Message: []interface{}{
								NodeText{
									Text: "Hello\\!",
								},
							},
						},
					},
				},
			},
			expected: "match {$userName :hasCase} " +
				"when 0 vocative {Hello, {$userName :person case=vocative format=printf type=string}\\!} " +
				"when 1 accusative {Please welcome {$userName :person case=accusative format=printf type=string}\\!} " +
				"when * neutral {Hello\\!}",
		},
		{
			name: "match, variants with no variables",
			input: []interface{}{
				NodeMatch{
					Selectors: []NodeExpr{
						{
							Value:    NodeVariable{Name: "count"},
							Function: NodeFunction{Name: "number"},
						},
					},
					Variants: []NodeVariant{
						{
							Keys: []string{"1"},
							Message: []interface{}{
								NodeText{Text: "Il y a "},
								NodeExpr{
									Function: NodeFunction{
										Name: "Placeholder",
										Options: []NodeOption{
											{Name: "format", Value: "printf"},
											{Name: "type", Value: "int"},
										},
									},
								},
								NodeText{Text: " pomme."},
							},
						},
						{
							Keys: []string{"*"},
							Message: []interface{}{
								NodeText{Text: "Il y a "},
								NodeExpr{
									Function: NodeFunction{
										Name: "Placeholder",
										Options: []NodeOption{
											{Name: "format", Value: "printf"},
											{Name: "type", Value: "int"},
										},
									},
								},
								NodeText{Text: " pommes."},
							},
						},
					},
				},
			},
			expected: "match {$count :number} " +
				"when 1 {Il y a {:Placeholder format=printf type=int} pomme.} " +
				"when * {Il y a {:Placeholder format=printf type=int} pommes.}",
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, err := Compile(test.input)

			require.Equal(t, test.err, err)
			assert.Equal(t, test.expected, l)
		})
	}
}
