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
