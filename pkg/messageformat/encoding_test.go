package messageformat

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MarshalText(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		expectedErr error

		name     string
		expected []byte
		input    AST
	}{
		{
			name:        "error, empty node expression",
			input:       AST{NodeExpr{}},
			expectedErr: errors.New("expression node must not be empty"),
		},
		{
			name: "error, no selectors",
			input: AST{
				NodeMatch{
					Variants: []NodeVariant{
						{
							Keys: []string{"1"},
						},
					},
				},
			},
			expectedErr: errors.New("there must be at least one selector"),
		},
		{
			name: "error, mismatching number of keys and selectors",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "one"}, Function: NodeFunction{Name: "func"}}},
					Variants: []NodeVariant{
						{
							Keys: []string{"1", "2"},
						},
					},
				},
			},
			expectedErr: errors.New("number of keys '2' for variant #0 don't match number of match selectors '1'"),
		},
		{
			name:        "no nodes",
			input:       AST{},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:     "single text node",
			input:    AST{NodeText{Text: "Hello, World\\!"}},
			expected: []byte("{Hello, World\\!}"),
		},
		{
			name: "multiple text nodes",
			input: AST{
				NodeText{Text: "Hello,"},
				NodeText{Text: " "},
				NodeText{Text: "World"},
				NodeText{Text: "\\!"},
			},
			expected: []byte("{Hello, World\\!}"),
		},
		{
			name:     "text with special characters",
			input:    AST{NodeText{Text: "\\{\\}\\|\\!\\@\\#\\%\\*\\<\\>\\/\\?\\~\\\\"}},
			expected: []byte("{\\{\\}\\|\\!\\@\\#\\%\\*\\<\\>\\/\\?\\~\\\\}"),
		},
		{
			name:     "text contains plus sign",
			input:    AST{NodeText{Text: "+ vēl \\%s"}},
			expected: []byte("{+ vēl \\%s}"),
		},
		{
			name:     "text contains minus sign",
			input:    AST{NodeText{Text: "- vēl \\%s"}},
			expected: []byte("{- vēl \\%s}"),
		},
		{
			name: "message with placeholder",
			input: AST{
				NodeText{Text: "Hello "},
				NodeExpr{Value: NodeVariable{Name: "name"}},
				NodeText{Text: ", your card expires on "},
				NodeExpr{
					Value: NodeVariable{Name: "exp"},
					Function: NodeFunction{
						Name: "datetime",
						Options: []NodeOption{
							{Name: "skeleton", Value: "yMMMdE"},
						},
					},
				},
				NodeText{Text: "!"},
			},
			expected: []byte("{Hello {$name}, your card expires on {$exp :datetime skeleton=yMMMdE}!}"),
		},
		{
			name: "match, single variant",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []Node{NodeText{Text: "Hello, world\\!"}}},
					},
				},
			},
			expected: []byte("match {$count} when * {Hello, world\\!}"),
		},
		{
			name: "match, single variant with function",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"*"}, Message: []Node{NodeText{Text: "Hello, world\\!"}}},
					},
				},
			},
			expected: []byte("match {$count :number} when * {Hello, world\\!}"),
		},
		{
			name: "match, multiple variants",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []Node{NodeText{Text: "Hello, friend\\!"}}},
						{Keys: []string{"*"}, Message: []Node{NodeText{Text: "Hello, friends\\!"}}},
					},
				},
			},
			expected: []byte("match {$count :number} when 1 {Hello, friend\\!} when * {Hello, friends\\!}"),
		},
		{
			name: "match, multiple variants 2",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"1"}, Message: []Node{NodeText{Text: "Buy one \\\\ apple!"}}},
						{Keys: []string{"*"}, Message: []Node{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeText{Text: " apples\\!"},
						}},
					},
				},
			},
			expected: []byte("match {$count :number} when 1 {Buy one \\\\ apple!} when * {Buy {$count} apples\\!}"),
		},
		{
			name: "match, variant with two variables",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
					Variants: []NodeVariant{
						{Keys: []string{"0"}, Message: []Node{NodeText{Text: "No apples\\!"}}},
						{Keys: []string{"1"}, Message: []Node{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeVariable{Name: "counts"},
							NodeText{Text: " apple\\!"},
						}},
						{Keys: []string{"*"}, Message: []Node{
							NodeText{Text: "Buy "},
							NodeVariable{Name: "count"},
							NodeText{Text: " apples 2\\!"},
						}},
					},
				},
			},
			expected: []byte("match {$count :number} " +
				"when 0 {No apples\\!} " +
				"when 1 {Buy {$count}{$counts} apple\\!} " +
				"when * {Buy {$count} apples 2\\!}"),
		},
		{
			name: "match, three variants, two selectors",
			input: AST{
				NodeMatch{
					Selectors: []NodeExpr{
						{Value: NodeVariable{Name: "userName"}, Function: NodeFunction{Name: "hasCase"}},
						{Value: NodeVariable{Name: "userLastName"}, Function: NodeFunction{Name: "hasCase"}},
					},
					Variants: []NodeVariant{
						{
							Keys: []string{"0", "vocative"},
							Message: []Node{
								NodeText{Text: "Hello, "},
								NodeExpr{
									Value: NodeVariable{Name: "userName"},
									Function: NodeFunction{
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
							Message: []Node{
								NodeText{Text: "Please welcome "},
								NodeExpr{
									Value: NodeVariable{Name: "userName"},
									Function: NodeFunction{
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
							Message: []Node{
								NodeText{
									Text: "Hello ",
								},
								NodeExpr{
									Value: NodeVariable{Name: "userLastName"},
									Function: NodeFunction{
										Name: "person",
										Options: []NodeOption{
											{Name: "case", Value: "neutral"},
											{Name: "format", Value: "printf"},
											{Name: "type", Value: "string"},
										},
									},
								},
								NodeText{
									Text: "\\.",
								},
							},
						},
					},
				},
			},
			expected: []byte("match {$userName :hasCase} {$userLastName :hasCase} " +
				"when 0 vocative {Hello, {$userName :person case=vocative format=printf type=string}\\!} " +
				"when 1 accusative {Please welcome {$userName :person case=accusative format=printf type=string}\\!} " +
				"when * neutral {Hello {$userLastName :person case=neutral format=printf type=string}\\.}"),
		},
		{
			name: "match, variants with no variables",
			input: AST{
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
							Message: []Node{
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
							Message: []Node{
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
			expected: []byte("match {$count :number} " +
				"when 1 {Il y a {:Placeholder format=printf type=int} pomme.} " +
				"when * {Il y a {:Placeholder format=printf type=int} pommes.}"),
		},
		{
			name: "expression is first followed by text",
			input: AST{
				NodeExpr{
					Function: NodeFunction{
						Name: "person", Options: []NodeOption{
							{Name: "gender", Value: "male"},
						},
					},
				},
				NodeText{Text: " hello "},
			},
			expected: []byte("{{:person gender=male} hello }"),
		},
		{
			name: "with placeholder at the end",
			input: AST{
				NodeExpr{
					Value: nil,
					Function: NodeFunction{
						Name: "Placeholder",
						Options: []NodeOption{
							{Name: "format", Value: "pythonVar"},
							{Name: "name", Value: "message"},
							{Name: "type", Value: "string"},
						},
					},
				},
				NodeText{Text: "\nThis may be triggered by: \n"},
				NodeExpr{
					Value: nil,
					Function: NodeFunction{
						Name: "Placeholder",
						Options: []NodeOption{
							{Name: "format", Value: "pythonVar"},
							{Name: "name", Value: "issues"},
							{Name: "type", Value: "string"},
						},
					},
				},
			},
			expected: []byte(`{{:Placeholder format=pythonVar name=message type=string}
			This may be triggered by:
			{:Placeholder format=pythonVar name=issues type=string}}`),
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, err := test.input.MarshalText()

			if test.expectedErr != nil {
				require.ErrorContains(t, err, test.expectedErr.Error())
				return
			}

			assert.Equal(t, string(test.expected), string(l))
		})
	}
}

//nolint:lll
func Test_UnmarshalText(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		input    []byte
		expected AST
	}{
		// singular tests
		{
			name:     "empty",
			input:    []byte(""),
			expected: AST(nil),
		},
		{
			name:     "text",
			input:    []byte("{Hello, World}"),
			expected: AST{NodeText{Text: "Hello, World"}},
		},
		{
			name:     "text with escaped curly braces",
			input:    []byte("{Hello, \\{World\\}}"),
			expected: AST{NodeText{Text: "Hello, \\{World\\}"}},
		},
		{
			name:  "text with variable",
			input: []byte("{Hello {$var} World}"),
			expected: AST{
				NodeText{Text: "Hello "},
				NodeExpr{Value: NodeVariable{Name: "var"}},
				NodeText{Text: " World"},
			},
		},
		{
			name:  "text with function",
			input: []byte("{Hello {:func} World}"),
			expected: AST{
				NodeText{Text: "Hello "},
				NodeExpr{Function: NodeFunction{Name: "func"}},
				NodeText{Text: " World"},
			},
		},
		{
			name:  "extracted placeholder pythonVar",
			input: []byte("{{:Placeholder name=object format=pythonVar type=string} does not exist in this database.}"),
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
			input: []byte("{{:Placeholder format=printf type=string} does not exist in {:Placeholder format=printf type=int}. database.}}"),
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
			input: []byte("match {$count} when * {Hello, world\\!}"),
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
			input: []byte("match {$count :number} when * {Hello, world\\!}"),
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
			input: []byte("match {$count :number} when 1 {Hello, friend\\!} when * {Hello, friends\\!} "),
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
			input: []byte("match {$count :number} when 1 {Buy one \\\\ apple\\!} when * {Buy {$count} apples\\!} "),
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
			input: []byte("match {$count :number} " +
				"when 0 {No apples\\!} " +
				"when 1 {Buy {$count}{$counts} apple\\!} " +
				"when * {Buy {$count} apples 2\\!} "),
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
			input: []byte("match {$count :number} " +
				"when 1 {Were having trouble loading this visualization. Queries are set to timeout after {:Placeholder name=sec format=bracketVar} second.}" +
				"when * {Were having trouble loading this visualization. Queries are set to timeout after {:Placeholder name=sec format=bracketVar} seconds.}"),
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

			var ast AST

			require.NoError(t, ast.UnmarshalText(test.input))

			assert.Equal(t, test.expected, ast)
		})
	}
}
