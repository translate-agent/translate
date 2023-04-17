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

			tree = append(tree, text)
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

func (p *parser) parseText() (NodeText, error) {
	if p.tokens[p.pos].Type != TokenTypePlaceholderOpen {
		return NodeText{}, errors.New(`text does not start with "{"`)
	}

	var text NodeText

	for {
		p.pos++

		if len(p.tokens) <= p.pos {
			return NodeText{}, errors.New("invalid text node")
		}

		token := p.tokens[p.pos]

		switch token.Type {
		case TokenTypeText:
			text.Text = token.Value
		// TODO
		// case TokenTypePlaceholderOpen:
		// 	p.parseExpr()
		case TokenTypePlaceholderClose:
			return text, nil
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
		case TokenTypePlaceholderClose:
			p.pos++

			return expr, nil
		}
	}
}

func (p *parser) parseVariant() (NodeVariant, error) {
	var variant NodeVariant

	// when * {Hello, world!}

	literal := p.nextToken()

	if literal.Type != TokenTypeLiteral {
		return NodeVariant{}, fmt.Errorf(`literal does not follow keword "when"`)
	}

	variant.Keys = append(variant.Keys, literal.Value)

	text, err := p.parseText()
	if err != nil {
		return NodeVariant{}, err
	}

	variant.Message = append(variant.Message, text)

	return variant, nil
}

func Parse(s string) ([]interface{}, error) {
	var p parser

	tree, err := p.parse(s)
	if err != nil {
		return nil, err
	}

	return tree, nil
}
