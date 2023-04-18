package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {

	for _, test := range []struct {
		name, message string
		expectedErr   bool
		expectedTree  []interface{}
	}{{
		name:         "text node",
		message:      "{Hello, world!}",
		expectedTree: []interface{}{NodeText{Text: "Hello, world!"}},
	}, {
		name:        "missing opening for text node",
		message:     "Hello, world!}",
		expectedErr: true,
	}, {
		name:        "missing closing for text node",
		message:     "{Hello, world!",
		expectedErr: true,
	}, {
		name:    "match",
		message: "match {$count} when * {Hello, world!}",
		expectedTree: []interface{}{
			NodeMatch{
				Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}}},
				Variants: []NodeVariant{
					{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world!"}}},
				},
			},
		},
	}, {
		name:        "variable starts with space",
		message:     "match {$ count} when * {Hello, world!}",
		expectedErr: true,
	}, {
		name:    "match with function",
		message: "match {$count :number} when * {Hello, world!}",
		expectedTree: []interface{}{
			NodeMatch{
				Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
				Variants: []NodeVariant{
					{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, world!"}}},
				},
			},
		},
	}, {
		name:        "function does not start with variable",
		message:     "match {:number} when * {Hello, world!}",
		expectedErr: true,
	}, {
		name:    "match with multiple variants",
		message: "match {$count :number} when 1 {Hello, friend!} when * {Hello, friends!} ",
		expectedTree: []interface{}{
			NodeMatch{
				Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
				Variants: []NodeVariant{
					{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Hello, friend!"}}},
					{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Hello, friends!"}}},
				},
			},
		},
	}, {
		name:    "match with plurals",
		message: "match {$count :number} when 1 {Buy one apple!} when * {Buy {$count} apples!} ",
		expectedTree: []interface{}{
			NodeMatch{
				Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
				Variants: []NodeVariant{
					{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "Buy one apple!"}}},
					{Keys: []string{"*"}, Message: []interface{}{
						NodeText{Text: "Buy "},
						NodeVariable{Name: "count"},
						NodeText{Text: " apples!"}}},
				},
			},
		},
	}, {
		name:    "match with plurals",
		message: "match {$count :number} when 0 {No apples!} when 1 {Buy {$count}{$count2} apple!} when * {Buy {$count} apples 2!} ",
		expectedTree: []interface{}{
			NodeMatch{
				Selectors: []NodeExpr{{Value: NodeVariable{Name: "count"}, Function: NodeFunction{Name: "number"}}},
				Variants: []NodeVariant{
					{Keys: []string{"0"}, Message: []interface{}{NodeText{Text: "No apples!"}}},
					{Keys: []string{"1"}, Message: []interface{}{
						NodeText{Text: "Buy "},
						NodeVariable{Name: "count"},
						NodeVariable{Name: "count2"},
						NodeText{Text: " apple!"}}},
					{Keys: []string{"*"}, Message: []interface{}{
						NodeText{Text: "Buy "},
						NodeVariable{Name: "count"},
						NodeText{Text: " apples 2!"}}},
				},
			},
		},
	}} {
		test := test

		t.Run(test.name, func(t *testing.T) {
			tree, err := Parse(test.message)

			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expectedTree, tree)
		})
	}
}
