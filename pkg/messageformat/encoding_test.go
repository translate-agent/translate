package messageformat

import (
	"errors"
	"fmt"
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
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world\\!"}}},
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
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world\\!"}}},
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
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Hello, friend\\!"}}},
						{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, friends\\!"}}},
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
						{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Buy one \\\\ apple!"}}},
						{Keys: []string{"*"}, Message: []interface{}{
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
									Text: "Hello ",
								},
								NodeExpr{
									NodeVariable{Name: "userLastName"},
									NodeFunction{
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
			expected: []byte("match {$count :number} " +
				"when 1 {Il y a {:Placeholder format=printf type=int} pomme.} " +
				"when * {Il y a {:Placeholder format=printf type=int} pommes.}"),
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

			assert.Equal(t, test.expected, l)
		})
	}
}

func Test_UnmarshalText(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		input       []byte
		expectedErr error
		expected    AST
	}{
		{
			name:     "empty",
			input:    []byte(""),
			expected: nil,
		},
		{
			name:     "empty expr",
			input:    []byte("{}"),
			expected: nil,
		},
		{
			name:     "expr with text",
			input:    []byte("{Hello, World!}"),
			expected: AST{NodeText{Text: "Hello, World!"}},
		},
		{
			name:  "match",
			input: []byte("match {$count} when * {Hello, world!}"),
			expected: AST{
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
			input: []byte("match {$count :number} when * {Hello, world!}"),
			expected: AST{
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
			input: []byte("match {$count :number} when 1 {Hello, friend!} when * {Hello, friends!} "),
			expected: AST{
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
			input: []byte("match {$count :number} when 1 {Buy one \\\\ apple!} when * {Buy {$count} apples!} "),
			expected: AST{
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
			input: []byte("match {$count :number} " +
				"when 0 {No apples!} " +
				"when 1 {Buy {$count}{$counts} apple!} " +
				"when * {Buy {$count} apples 2!} "),
			expected: AST{
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
			input:       []byte("match $count :number} "),
			expectedErr: fmt.Errorf("expression does not start with \"{\""),
		},
		{
			name:     "input with curly braces in it",
			input:    []byte(`{Chart [\{\}] was added to dashboard [\{\}]}`),
			expected: AST{NodeText{Text: "Chart [{}] was added to dashboard [{}]"}},
		},
		{
			name:     "input with plus sign in it ",
			input:    []byte(`{+ vēl %s}`),
			expected: AST{NodeText{Text: "+ vēl %s"}},
		},
		{
			name:     "input with minus sign in it ",
			input:    []byte(`{- vēl %s}`),
			expected: AST{NodeText{Text: "- vēl %s"}},
		},
	} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var ast AST

			err := ast.UnmarshalText(test.input)

			if test.expectedErr != nil {
				require.Errorf(t, err, test.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, ast)
		})
	}
}
