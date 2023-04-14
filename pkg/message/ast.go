package message

import (
	"errors"
	"strings"
)

type nodeMatch struct {
	Selectors []nodeExpr
	Variants  []nodeVariant
}

type nodeVariable struct {
	Name string
}

type nodeText struct {
	Text string
}

type nodeExpr struct {
	Value    interface{}
	Function nodeFunction
}

type nodeVariant struct {
	Keys    []string
	Message []interface{}
}

type nodeFunction struct {
	Name string
}

func TokensToMessageFormat(tokens []Token) (nodeMatch, error) {
	if len(tokens) == 0 {
		return nodeMatch{}, errors.New("tokens[] is empty")
	}

	var (
		match               nodeMatch
		currentVariant      nodeVariant
		currentExpr         nodeExpr
		isPlaceholderClosed bool
		currentLevel        int
		token               Token
	)

	// If message format starts with placeholder {
	if tokens[0].Type == PlaceholderOpen {
		tokens[0] = token
		tokens[len(tokens)-1] = token
	}

	for i, token := range tokens {
		switch token.Type {
		case Literal:
			currentVariant.Keys = append(currentVariant.Keys, strings.TrimSpace(token.Value))
			currentLevel++
		case Text:
			currentVariant.Message = append(currentVariant.Message, nodeText{Text: token.Value})
		case Variable:
			switch token.Level {
			case 1:
				currentExpr.Value = nodeVariable{Name: strings.TrimSpace(token.Value)}
				if tokens[i+1].Type == Function {
					currentExpr.Function = nodeFunction{Name: tokens[i+1].Value}
				}

				match.Selectors = append(match.Selectors, currentExpr)
				currentExpr = nodeExpr{}
			default:
				currentVariant.Message = append(currentVariant.Message, nodeVariable{Name: token.Value})
			}
		case Function:
			currentExpr.Function = nodeFunction{Name: token.Value}
		case PlaceholderOpen:
			isPlaceholderClosed = false
		case PlaceholderClose:
			if token.Level == currentLevel && token.Level != 0 {
				match.Variants = append(match.Variants, currentVariant)
				currentVariant = nodeVariant{}
				currentLevel = 0
			}

			isPlaceholderClosed = true
		case Keyword:
			continue
		}
	}

	if !isPlaceholderClosed {
		return nodeMatch{}, errors.New("placeholder is not closed")
	}

	if tokens[len(tokens)-1] == token {
		match.Variants = append(match.Variants, currentVariant)
	}

	return match, nil
}
