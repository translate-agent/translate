package messageformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:lll
func Test_Parse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		input    string
		expected AST
	}{
		// singular tests
		{
			name:     "empty",
			input:    "",
			expected: AST(nil),
		},
		{
			name:     "text",
			input:    "{{{{Hello, World}}}}",
			expected: AST{NodeText{Text: "Hello, World"}},
		},
		{
			name:     "text with escaped curly braces",
			input:    "{{{{Hello, \\{World\\}}}}",
			expected: AST{NodeText{Text: "Hello, \\{World\\}"}},
		},
		{
			name:  "text with variable",
			input: "{{{{Hello {$var} World}}}}",
			expected: AST{
				NodeText{Text: "Hello "},
				NodeExpr{Value: NodeVariable{Name: "var"}},
				NodeText{Text: " World"},
			},
		},
		{
			name:  "text with function",
			input: "{{{{Hello {:func} World}}}}",
			expected: AST{
				NodeText{Text: "Hello "},
				NodeExpr{Function: NodeFunction{Name: "func"}},
				NodeText{Text: " World"},
			},
		},
		{
			name:  "extracted placeholder pythonVar",
			input: "{{{{{:Placeholder name=object format=pythonVar type=string} does not exist in this database.}}}}}",
			expected: AST{
				NodeExpr{
					Value: nil,
					Function: NodeFunction{
						Name: "Placeholder",
						Options: []NodeOption{
							{Name: "name", Value: "object"},
							{Name: "format", Value: "pythonVar"},
							{Name: "type", Value: "string"},
						},
					},
				},
				NodeText{Text: " does not exist in this database."},
			},
		},
		{
			name:  "extracted placeholders printf style",
			input: "{{{{{:Placeholder format=printf type=string} does not exist in {:Placeholder format=printf type=int}. database.}}}}}",
			expected: AST{
				NodeExpr{
					Value: nil,
					Function: NodeFunction{
						Name: "Placeholder",
						Options: []NodeOption{
							{Name: "format", Value: "printf"},
							{Name: "type", Value: "string"},
						},
					},
				},
				NodeText{Text: " does not exist in "},
				NodeExpr{
					Value: nil,
					Function: NodeFunction{
						Name: "Placeholder",
						Options: []NodeOption{
							{Name: "format", Value: "printf"},
							{Name: "type", Value: "int"},
						},
					},
				},
				NodeText{Text: ". database."},
			},
		},
		// plural tests
		{
			name:  "single match",
			input: "{{match {$count} when * {{Hello, world\\!}}}}",
			expected: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []Node{NodeText{Text: "Hello, world\\!"}}},
					},
				},
			},
		},
		{
			name:  "single match with function",
			input: "{{match {$count :number} when * {{Hello, world\\!}}}}",
			expected: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []Node{NodeText{Text: "Hello, world\\!"}}},
					},
				},
			},
		},
		{
			name:  "match with multiple variants",
			input: "{{match {$count :number} when 1 {{Hello, friend\\!}} when * {{Hello, friends\\!}}}} ",
			expected: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []Node{NodeText{Text: "Hello, friend\\!"}}},
						{Keys: []string{"*"}, Message: []Node{NodeText{Text: "Hello, friends\\!"}}},
					},
				},
			},
		},
		{
			name:  "match with plurals",
			input: "{{match {$count :number} when 1 {{Buy one \\\\ apple\\!} when * {{Buy {$count} apples\\!}}}}",
			expected: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []Node{NodeText{Text: "Buy one \\\\ apple\\!"}}},
						{Keys: []string{"*"}, Message: []Node{
							NodeText{Text: "Buy "},
							NodeExpr{Value: NodeVariable{Name: "count"}},
							NodeText{Text: " apples\\!"},
						}},
					},
				},
			},
		},
		{
			name: "match with two variables in variant",
			input: "{{match {$count :number} " +
				"when 0 {{No apples\\!}} " +
				"when 1 {{Buy {$count}{$counts} apple\\!}} " +
				"when * {{Buy {$count} apples 2\\!}}}} ",
			expected: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"0"}, Message: []Node{NodeText{Text: "No apples\\!"}}},
						{Keys: []string{"1"}, Message: []Node{
							NodeText{Text: "Buy "},
							NodeExpr{Value: NodeVariable{Name: "count"}},
							NodeExpr{Value: NodeVariable{Name: "counts"}},
							NodeText{Text: " apple\\!"},
						}},
						{Keys: []string{"*"}, Message: []Node{
							NodeText{Text: "Buy "},
							NodeExpr{Value: NodeVariable{Name: "count"}},
							NodeText{Text: " apples 2\\!"},
						}},
					},
				},
			},
		},
		{
			name: "match plural with extracted bracketVar placeholder",
			input: "{{match {$count :number} " +
				"when 1 {{Were having trouble loading this visualization. Queries are set to timeout after {:Placeholder name=sec format=bracketVar} second.}}" +
				"when * {{Were having trouble loading this visualization. Queries are set to timeout after {:Placeholder name=sec format=bracketVar} seconds.}}}}",
			expected: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{
							Keys: []string{"1"},
							Message: []Node{
								NodeText{Text: "Were having trouble loading this visualization. Queries are set to timeout after "},
								NodeExpr{
									Value: nil,
									Function: NodeFunction{
										Name: "Placeholder",
										Options: []NodeOption{
											{Name: "name", Value: "sec"},
											{Name: "format", Value: "bracketVar"},
										},
									},
								},
								NodeText{Text: " second."},
							},
						},
						{
							Keys: []string{"*"},
							Message: []Node{
								NodeText{Text: "Were having trouble loading this visualization. Queries are set to timeout after "},
								NodeExpr{
									Value: nil,
									Function: NodeFunction{
										Name: "Placeholder",
										Options: []NodeOption{
											{Name: "name", Value: "sec"},
											{Name: "format", Value: "bracketVar"},
										},
									},
								},
								NodeText{Text: " seconds."},
							},
						},
					},
				},
			},
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, err := Parse(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.expected, l)
		})
	}
}
