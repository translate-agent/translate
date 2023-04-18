package message

import (
	"errors"
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
	Value    interface{} // NodeLiteral | NodeVariable
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
	if len(tokens) == 0 {
		return NodeMatch{}, errors.New("tokens[] is empty")
	}

	var (
		match               NodeMatch
		currentVariant      NodeVariant
		currentExpr         NodeExpr
		isPlaceholderClosed bool
		currentLevel        int
		token               Token
	)

	// If message format starts with placeholder {
	if tokens[0].Type == TokenTypeDelimiterOpen {
		tokens[0] = token
		tokens[len(tokens)-1] = token
	}

	for i, token := range tokens {
		switch token.Type {
		case TokenTypeLiteral:
			currentVariant.Keys = append(currentVariant.Keys, strings.TrimSpace(token.Value))
			currentLevel++
		case TokenTypeText:
			currentVariant.Message = append(currentVariant.Message, NodeText{Text: token.Value})
		case TokenTypeVariable:
			switch token.Level {
			case 1:
				currentExpr.Value = NodeVariable{Name: strings.TrimSpace(token.Value)}
				if tokens[i+1].Type == TokenTypeFunction {
					currentExpr.Function = NodeFunction{Name: tokens[i+1].Value}
				}

				match.Selectors = append(match.Selectors, currentExpr)
				currentExpr = NodeExpr{}
			default:
				currentVariant.Message = append(currentVariant.Message, NodeVariable{Name: token.Value})
			}
		case TokenTypeFunction:
			currentExpr.Function = NodeFunction{Name: token.Value}
		case TokenTypeDelimiterOpen:
			isPlaceholderClosed = false
		case TokenTypeDelimiterClose:
			if token.Level == currentLevel && token.Level != 0 {
				match.Variants = append(match.Variants, currentVariant)
				currentVariant = NodeVariant{}
				currentLevel = 0
			}

			isPlaceholderClosed = true
		case TokenTypeKeyword:
			continue
		}
	}

	if !isPlaceholderClosed {
		return NodeMatch{}, errors.New("placeholder is not closed")
	}

	if tokens[len(tokens)-1] == token {
		match.Variants = append(match.Variants, currentVariant)
	}

	return match, nil
}
