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
		expected    NodeMatch
		expectedErr error
		input       []Token
	}{
		{
			name: "message with variable",
			input: []Token{
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "Hello, ", Level: 1},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 2},
				{Type: TokenTypeVariable, Value: "userName", Level: 2},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 2},
				{Type: TokenTypeText, Value: "!", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
			},
			expected: NodeMatch{
				Variants: []NodeVariant{
					{
						Keys: []string(nil),
						Message: []interface{}{
							NodeText{
								Text: "Hello, ",
							},
							NodeVariable{
								Name: "userName",
							},
							NodeText{
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
				{Type: TokenTypeKeyword, Value: "match", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeVariable, Value: "count", Level: 1},
				{Type: TokenTypeFunction, Value: "number", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: " when", Level: 0},
				{Type: TokenTypeLiteral, Value: "1", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have one notification.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "*", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have ", Level: 1},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 2},
				{Type: TokenTypeVariable, Value: "count", Level: 2},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 2},
				{Type: TokenTypeText, Value: " notifications.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
			},
			expected: NodeMatch{
				Selectors: []NodeExpr{
					{
						Value: NodeVariable{
							Name: "count",
						},
						Function: NodeFunction{
							Name: "number",
						},
					},
				},
				Variants: []NodeVariant{
					{
						Keys: []string{"1"},
						Message: []interface{}{
							NodeText{
								Text: "You have one notification.",
							},
						},
					},
					{
						Keys: []string{"*"},
						Message: []interface{}{
							NodeText{
								Text: "You have ",
							},
							NodeVariable{
								Name: "count",
							},
							NodeText{
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
				{Type: TokenTypeKeyword, Value: "match", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeVariable, Value: "count", Level: 1},
				{Type: TokenTypeFunction, Value: "number", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "0", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have no notifications.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "1", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have one notification.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "*", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have ", Level: 1},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 2},
				{Type: TokenTypeVariable, Value: "count", Level: 2},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 2},
				{Type: TokenTypeText, Value: " notifications.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
			},
			expected: NodeMatch{
				Selectors: []NodeExpr{
					{
						Value: NodeVariable{
							Name: "count",
						},
						Function: NodeFunction{
							Name: "number",
						},
					},
				},
				Variants: []NodeVariant{
					{
						Keys: []string{"0"},
						Message: []interface{}{
							NodeText{
								Text: "You have no notifications.",
							},
						},
					},
					{
						Keys: []string{"1"},
						Message: []interface{}{
							NodeText{
								Text: "You have one notification.",
							},
						},
					},
					{
						Keys: []string{"*"},
						Message: []interface{}{
							NodeText{
								Text: "You have ",
							},
							NodeVariable{
								Name: "count",
							},
							NodeText{
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
				{Type: TokenTypeKeyword, Value: "match", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeVariable, Value: "count", Level: 1},
				{Type: TokenTypeFunction, Value: "number", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeVariable, Value: "num ", Level: 1},
				{Type: TokenTypeFunction, Value: "number", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "0", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have no notifications.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "1", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have one notification.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
				{Type: TokenTypeKeyword, Value: "when", Level: 0},
				{Type: TokenTypeLiteral, Value: "*", Level: 0},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "You have ", Level: 1},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 2},
				{Type: TokenTypeVariable, Value: "count", Level: 2},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 2},
				{Type: TokenTypeText, Value: " notifications.", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
			},
			expected: NodeMatch{
				Selectors: []NodeExpr{
					{
						Value: NodeVariable{
							Name: "count",
						},
						Function: NodeFunction{
							Name: "number",
						},
					},
					{
						Value: NodeVariable{
							Name: "num",
						},
						Function: NodeFunction{
							Name: "number",
						},
					},
				},
				Variants: []NodeVariant{
					{
						Keys: []string{"0"},
						Message: []interface{}{
							NodeText{
								Text: "You have no notifications.",
							},
						},
					},
					{
						Keys: []string{"1"},
						Message: []interface{}{
							NodeText{
								Text: "You have one notification.",
							},
						},
					},
					{
						Keys: []string{"*"},
						Message: []interface{}{
							NodeText{
								Text: "You have ",
							},
							NodeVariable{
								Name: "count",
							},
							NodeText{
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
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 1},
				{Type: TokenTypeText, Value: "Hello, ", Level: 1},
				{Type: TokenTypeDelimiterOpen, Value: "{", Level: 2},
				{Type: TokenTypeVariable, Value: "userName", Level: 2},
				{Type: TokenTypeText, Value: "!", Level: 1},
				{Type: TokenTypeDelimiterClose, Value: "}", Level: 1},
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
