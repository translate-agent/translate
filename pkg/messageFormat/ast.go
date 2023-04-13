package messageFormat

import (
	"fmt"
	"strings"
)

type NodeMatch struct {
	Selectors []NodeExpr
	Variants  []NodeVariant
}

type NodeVariable struct {
	Name string
}

type NodeLiteral struct {
	Value string
}

type NodeText struct {
	Text string
}

type NodeExpr struct {
	Value    interface{} // NodeVariable | NodeLiteral
	Function NodeFunction
}

type NodeVariant struct {
	Keys    []string
	Message []interface{}
}

type NodeFunction struct {
	Name string
}

func TokensToMessageFormat(tokens []Token) (NodeMatch, error) {
	var (
		nodeMatch           NodeMatch
		currentNodeVariant  NodeVariant
		currentNodeExpr     NodeExpr
		isPlaceholderClosed bool
		currentLevel        int
		emptyToken          Token
	)

	// If message format starts with placeholder {
	if tokens[0].Type == PlaceholderOpen {
		tokens[0] = emptyToken
		tokens[len(tokens)-1] = emptyToken
	}

	for i, token := range tokens {
		switch token.Type { //nolint:exhaustive
		case Literal:
			currentNodeVariant.Keys = append(currentNodeVariant.Keys, strings.TrimSpace(token.Value))
			currentLevel++
		case Text:
			currentNodeVariant.Message = append(currentNodeVariant.Message, NodeText{Text: token.Value})
		case Variable:
			switch token.Level {
			case 1:
				currentNodeExpr.Value = NodeVariable{Name: strings.TrimSpace(token.Value)}
				if tokens[i+1].Type == Function {
					currentNodeExpr.Function = NodeFunction{Name: tokens[i+1].Value}
				}

				nodeMatch.Selectors = append(nodeMatch.Selectors, currentNodeExpr)
				currentNodeExpr = NodeExpr{}
			default:
				currentNodeVariant.Message = append(currentNodeVariant.Message, NodeVariable{Name: token.Value})
			}
		case Function:
			currentNodeExpr.Function = NodeFunction{Name: token.Value}
		case PlaceholderOpen:
			isPlaceholderClosed = false
		case PlaceholderClose:
			if token.Level == currentLevel && token.Level != 0 {
				nodeMatch.Variants = append(nodeMatch.Variants, currentNodeVariant)
				currentNodeVariant = NodeVariant{}
				currentLevel = 0
			}

			isPlaceholderClosed = true
		default:
			continue
		}
	}

	if !isPlaceholderClosed {
		return NodeMatch{}, fmt.Errorf("placeholder is not closed")
	}

	if tokens[len(tokens)-1] == emptyToken {
		nodeMatch.Variants = append(nodeMatch.Variants, currentNodeVariant)
	}

	return nodeMatch, nil
}
