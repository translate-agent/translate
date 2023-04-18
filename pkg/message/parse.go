package message

import (
	"errors"
	"fmt"
)

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) parse(s string) ([]interface{}, error) {
	var err error

	if p.tokens, err = Lex(s); err != nil {
		return nil, err
	}

	var tree []interface{}

	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]

		switch {
		default:
			return nil, fmt.Errorf("unknown token: %+v", token)
		case token.Type == TokenTypeKeyword && token.Value == Match:
			match, err := p.parseMatch()
			if err != nil {
				return nil, err
			}

			tree = append(tree, match)
		case token.Type == TokenTypePlaceholderOpen:
			text, err := p.parseText()
			if err != nil {
				return nil, err
			}

			for _, element := range text {
				tree = append(tree, element)
			}

		}

		p.pos++
	}

	return tree, nil

}

func (p *parser) nextToken() Token {
	p.pos++
	return p.tokens[p.pos]
}

func (p *parser) lookup(pos int) Token {
	if len(p.tokens) <= pos {
		return Token{}
	}

	return p.tokens[pos]
}

// Old Version
//func (p *parser) parseText() (NodeText, error) {
//	if p.tokens[p.pos].Type != TokenTypePlaceholderOpen {
//		return NodeText{}, errors.New(`text does not start with "{"`)
//	}
//
//	var text NodeText
//
//	for {
//		p.pos++
//
//		if len(p.tokens) <= p.pos {
//			return NodeText{}, errors.New("invalid text node")
//		}
//
//		token := p.tokens[p.pos]
//
//		switch token.Type {
//		case TokenTypeText:
//			text.Text = token.Value
//		// TODO
//		/*		case TokenTypePlaceholderOpen:
//				//
//				//expr, err := p.parseExpr()
//				//if err != nil {
//				//	return NodeText{}, err
//				//}
//				//expr.Value = expr
//				p.pos++*/
//		case TokenTypePlaceholderClose:
//			return text, nil
//		}
//	}
//}

// New version
func (p *parser) parseText() ([]interface{}, error) {
	if p.tokens[p.pos].Type != TokenTypePlaceholderOpen {
		return nil, errors.New(`text does not start with "{"`)
	}

	var variantValues []interface{}
	//var text NodeText

	for {
		p.pos++

		if len(p.tokens) <= p.pos {
			return nil, errors.New("invalid text node")
		}

		token := p.tokens[p.pos]

		switch token.Type {
		case TokenTypeText:
			variantValues = append(variantValues, NodeText{Text: token.Value})
		// TODO
		case TokenTypePlaceholderOpen:
			variable, err := p.parseVariable()
			if err != nil {
				return nil, fmt.Errorf("new error: %w", err)
			}
			variantValues = append(variantValues, variable)
			p.pos++
		case TokenTypePlaceholderClose:
			return variantValues, nil
		}
	}
}

func (p *parser) parseMatch() (NodeMatch, error) {
	var match NodeMatch

	if p.tokens[p.pos].Type != TokenTypeKeyword || p.tokens[p.pos].Value != Match {
		return match, errors.New(`match node does not start with "match"`)
	}

	for {
		p.pos++

		if len(p.tokens) <= p.pos {
			// TODO: verify we have at least one variant
			return match, nil
		}

		token := p.tokens[p.pos]

		switch {
		case token.Type == TokenTypePlaceholderOpen:
			expr, err := p.parseExpr()
			if err != nil {
				return NodeMatch{}, err
			}

			match.Selectors = append(match.Selectors, expr)
		case token.Type == TokenTypeKeyword && token.Value == When:
			variant, err := p.parseVariant()
			if err != nil {
				return NodeMatch{}, err
			}

			match.Variants = append(match.Variants, variant)
		}
	}
}

func (p *parser) parseExpr() (NodeExpr, error) {
	// {$count}
	if p.tokens[p.pos].Type != TokenTypePlaceholderOpen {
		return NodeExpr{}, errors.New(`expression does not start with "{"`)
	}

	var expr NodeExpr

	for {
		p.pos++

		token := p.tokens[p.pos]

		switch token.Type {
		case TokenTypeVariable:
			expr.Value = NodeVariable{Name: token.Value}
		case TokenTypeFunction:
			function, err := p.parseFunction()
			if err != nil {
				return NodeExpr{}, err
			}

			expr.Function = function
		case TokenTypePlaceholderClose:
			//p.pos++

			return expr, nil

		}
	}
}

func (p *parser) parseVariant() (NodeVariant, error) {
	var variant NodeVariant

	// when * {Hello, world!}

	literal := p.nextToken()

	if literal.Type != TokenTypeLiteral {
		return NodeVariant{}, fmt.Errorf(`literal does not follow keyword "when"`)
	}

	variant.Keys = append(variant.Keys, literal.Value)

	p.pos++

	text, err := p.parseText()
	if err != nil {
		return NodeVariant{}, err
	}

	for _, element := range text {
		variant.Message = append(variant.Message, element)
	}

	return variant, nil
}

func (p *parser) parseFunction() (NodeFunction, error) {
	var function NodeFunction

	if p.tokens[p.pos-1].Type != TokenTypeVariable {
		return function, errors.New(`function does not follow variable`)
	}

	function.Name = p.tokens[p.pos].Value

	return function, nil
}

func (p *parser) parseVariable() (NodeVariable, error) {
	p.pos++

	var variable NodeVariable

	if p.tokens[p.pos-1].Type != TokenTypePlaceholderOpen {
		return variable, errors.New(`function does not follow placeholder open`)
	}

	variable.Name = p.tokens[p.pos].Value

	return variable, nil
}

func Parse(s string) ([]interface{}, error) {
	var p parser

	tree, err := p.parse(s)
	if err != nil {
		return nil, err
	}

	return tree, nil
}
