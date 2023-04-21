package message

import (
	"errors"
	"fmt"
)

type parser struct {
	tokens []Token
	pos    int
}

func Parse(s string) ([]interface{}, error) {
	var p parser

	tree, err := p.parse(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return tree, nil
}

func (p *parser) parse(s string) ([]interface{}, error) {
	var err error

	if p.tokens, err = Lex(s); err != nil {
		return nil, fmt.Errorf("failed to lex: %w", err)
	}

	var tree []interface{}

	for p.pos < len(p.tokens) {
		token := p.currentToken()

		switch {
		default:
			return nil, fmt.Errorf("unknown token: %+v", token)
		case token.Type == TokenTypeKeyword && token.Value == KeywordMatch:
			match, err := p.parseMatch()
			if err != nil {
				return nil, err
			}

			tree = append(tree, match)
		case token.Type == TokenTypeSeparatorOpen:
			text, err := p.parseText()
			if err != nil {
				return nil, err
			}

			tree = append(tree, text...)
		case token.Type == TokenTypeEOF:
			return tree, nil
		}

		p.pos++
	}

	return tree, nil
}

func (p *parser) parseText() ([]interface{}, error) {
	if p.currentToken().Type != TokenTypeSeparatorOpen {
		return nil, errors.New(`text does not start with "{"`)
	}

	var text []interface{}

	for !p.isEOF() {
		token := p.nextToken()

		switch token.Type {
		case TokenTypeText:
			text = append(text, NodeText{Text: token.Value})
		case TokenTypeSeparatorOpen:
			variable, err := p.parseVariable()
			if err != nil {
				return nil, fmt.Errorf("parse variable: %w", err)
			}

			text = append(text, variable)
			p.pos++
		case TokenTypeSeparatorClose:
			p.pos++

			return text, nil
		case TokenTypeUnknown, TokenTypeKeyword, TokenTypeLiteral, TokenTypeFunction, TokenTypeVariable, TokenTypeEOF:
		}
	}

	return nil, fmt.Errorf("invalid text node")
}

func (p *parser) parseMatch() (NodeMatch, error) {
	var match NodeMatch

	if v := p.currentToken(); v.Type != TokenTypeKeyword || v.Value != KeywordMatch {
		return match, errors.New(`match node does not start with "match"`)
	}

	for !p.isEOF() {
		token := p.currentToken()

		if token.Type != TokenTypeKeyword {
			return NodeMatch{}, fmt.Errorf("invalid match")
		}

		switch token.Value {
		case KeywordMatch:
			p.pos++

			expr, err := p.parseExpr()
			if err != nil {
				return NodeMatch{}, err
			}

			match.Selectors = append(match.Selectors, expr)
		case KeywordWhen:
			variant, err := p.parseVariant()
			if err != nil {
				return NodeMatch{}, fmt.Errorf("parse variant: %w", err)
			}

			match.Variants = append(match.Variants, variant)
		}
	}

	return match, nil
}

func (p *parser) parseExpr() (NodeExpr, error) {
	if p.currentToken().Type != TokenTypeSeparatorOpen {
		return NodeExpr{}, errors.New(`expression does not start with "{"`)
	}

	var expr NodeExpr

	for !p.isEOF() {
		token := p.nextToken()

		switch token.Type {
		case TokenTypeVariable:
			expr.Value = NodeVariable{Name: token.Value}
		case TokenTypeFunction:
			function, err := p.parseFunction()
			if err != nil {
				return NodeExpr{}, err
			}

			expr.Function = function
		case TokenTypeSeparatorClose:
			p.pos++

			return expr, nil
		case TokenTypeUnknown, TokenTypeKeyword, TokenTypeSeparatorOpen, TokenTypeLiteral, TokenTypeText, TokenTypeEOF:
		}
	}

	return NodeExpr{}, fmt.Errorf("invalid expression node")
}

func (p *parser) parseVariant() (NodeVariant, error) {
	var variant NodeVariant

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

	variant.Message = append(variant.Message, text...)

	return variant, nil
}

func (p *parser) parseFunction() (NodeFunction, error) {
	var function NodeFunction

	if p.tokens[p.pos-1].Type != TokenTypeVariable {
		return function, errors.New(`function does not follow variable`)
	}

	function.Name = p.currentToken().Value

	return function, nil
}

func (p *parser) parseVariable() (NodeVariable, error) {
	p.pos++

	var variable NodeVariable

	if p.tokens[p.pos-1].Type != TokenTypeSeparatorOpen {
		return variable, errors.New(`function does not follow placeholder open`)
	}

	variable.Name = p.currentToken().Value

	return variable, nil
}

func (p *parser) currentToken() Token {
	return p.tokens[p.pos]
}

func (p *parser) nextToken() Token {
	p.pos++

	return p.tokens[p.pos]
}

func (p *parser) isEOF() bool { return p.tokens[p.pos].Type == TokenTypeEOF }
