package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestTokensToMessageFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expected    nodeMatch
		expectedErr error
		input       []Token
	}{
		{
			name: "message with variable",
			input: []Token{
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "Hello, ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "userName", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: "!", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
			expected: nodeMatch{
				Variants: []nodeVariant{
					{
						Keys: []string(nil),
						Message: []interface{}{
							nodeText{
								Text: "Hello, ",
							},
							nodeVariable{
								Name: "userName",
							},
							nodeText{
								Text: "!",
							},
						},
					},
				},
			},
		},
		{
			name: "message format is plural",
			input: []Token{
				{Type: Keyword, Value: "match ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "count ", Level: 1},
				{Type: Function, Value: "number", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "1 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have one notification.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "* ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "count", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: " notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
			expected: nodeMatch{
				Selectors: []nodeExpr{
					{
						Value: nodeVariable{
							Name: "count",
						},
						Function: nodeFunction{
							Name: "number",
						},
					},
				},
				Variants: []nodeVariant{
					{
						Keys: []string{"1"},
						Message: []interface{}{
							nodeText{
								Text: "You have one notification.",
							},
						},
					},
					{
						Keys: []string{"*"},
						Message: []interface{}{
							nodeText{
								Text: "You have ",
							},
							nodeVariable{
								Name: "count",
							},
							nodeText{
								Text: " notifications.",
							},
						},
					},
				},
			},
		},
		{
			name: "message with multiple plurals",
			input: []Token{
				{Type: Keyword, Value: "match ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "count ", Level: 1},
				{Type: Function, Value: "number", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "0 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have no notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "1 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have one notification.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "* ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "count", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: " notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
			expected: nodeMatch{
				Selectors: []nodeExpr{
					{
						Value: nodeVariable{
							Name: "count",
						},
						Function: nodeFunction{
							Name: "number",
						},
					},
				},
				Variants: []nodeVariant{
					{
						Keys: []string{"0"},
						Message: []interface{}{
							nodeText{
								Text: "You have no notifications.",
							},
						},
					},
					{
						Keys: []string{"1"},
						Message: []interface{}{
							nodeText{
								Text: "You have one notification.",
							},
						},
					},
					{
						Keys: []string{"*"},
						Message: []interface{}{
							nodeText{
								Text: "You have ",
							},
							nodeVariable{
								Name: "count",
							},
							nodeText{
								Text: " notifications.",
							},
						},
					},
				},
			},
		},
		{
			name: "message with multiple plurals and selectors",
			input: []Token{
				{Type: Keyword, Value: "match ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "count ", Level: 1},
				{Type: Function, Value: "number", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Variable, Value: "num ", Level: 1},
				{Type: Function, Value: "number", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "0 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have no notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "1 ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have one notification.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
				{Type: Keyword, Value: " when ", Level: 0},
				{Type: Literal, Value: "* ", Level: 0},
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "You have ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "count", Level: 2},
				{Type: PlaceholderClose, Value: "}", Level: 2},
				{Type: Text, Value: " notifications.", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
			expected: nodeMatch{
				Selectors: []nodeExpr{
					{
						Value: nodeVariable{
							Name: "count",
						},
						Function: nodeFunction{
							Name: "number",
						},
					},
					{
						Value: nodeVariable{
							Name: "num",
						},
						Function: nodeFunction{
							Name: "number",
						},
					},
				},
				Variants: []nodeVariant{
					{
						Keys: []string{"0"},
						Message: []interface{}{
							nodeText{
								Text: "You have no notifications.",
							},
						},
					},
					{
						Keys: []string{"1"},
						Message: []interface{}{
							nodeText{
								Text: "You have one notification.",
							},
						},
					},
					{
						Keys: []string{"*"},
						Message: []interface{}{
							nodeText{
								Text: "You have ",
							},
							nodeVariable{
								Name: "count",
							},
							nodeText{
								Text: " notifications.",
							},
						},
					},
				},
			},
		},
		{
			name: "message with wrong format",
			input: []Token{
				{Type: PlaceholderOpen, Value: "{", Level: 1},
				{Type: Text, Value: "Hello, ", Level: 1},
				{Type: PlaceholderOpen, Value: "{", Level: 2},
				{Type: Variable, Value: "userName", Level: 2},
				{Type: Text, Value: "!", Level: 1},
				{Type: PlaceholderClose, Value: "}", Level: 1},
			},
			expectedErr: fmt.Errorf("placeholder is not closed"),
		},
		{
			name:        "message with wrong format",
			input:       []Token{},
			expectedErr: fmt.Errorf("tokens[] is empty"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := TokensToMessageFormat(tt.input)

			if tt.expectedErr != nil {
				assert.Errorf(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}
