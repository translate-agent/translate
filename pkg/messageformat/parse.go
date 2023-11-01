package messageformat

import (
	"errors"
	"fmt"
)

type parser struct {
	lex    *lexer
	text   string
	tokens []Token
	pos    int
}

func (p *parser) collect() {
	for {
		v := p.lex.nextToken()

		switch v.typ {
		case tokenTypeVariable, tokenTypeFunction, tokenTypeSeparatorOpen,
			tokenTypeSeparatorClose, tokenTypeText, tokenTypeOpeningFunction,
			tokenTypeClosingFunction, tokenTypeKeyword, tokenTypeLiteral, tokenTypeOption:
			p.tokens = append(p.tokens, v)
		case tokenTypeEOF:
			p.tokens = append(p.tokens, v)
			return
		case tokenTypeError:
			return
		}
	}
}

func Parse(text string) ([]interface{}, error) {
	var p parser
	p.text = text
	p.lex = lex(text)
	p.collect()

	tree, err := p.parse()
	if err != nil {
		return nil, fmt.Errorf("parse MessageFormat: %w", err)
	}

	return tree, nil
}

func (p *parser) parse() ([]interface{}, error) {
	var tree []interface{}

	for p.pos < len(p.tokens) {
		token := p.currentToken()

		switch {
		default:
			return nil, fmt.Errorf("unknown token: %+v", token)
		case token.typ == tokenTypeKeyword && token.val == KeywordMatch:
			match, err := p.parseMatch()
			if err != nil {
				return nil, fmt.Errorf("parse match: %w", err)
			}

			tree = append(tree, match)
		case token.typ == tokenTypeSeparatorOpen:
			text, err := p.parseText()
			if err != nil {
				return nil, fmt.Errorf("parse text: %w", err)
			}

			tree = append(tree, text...)
		case token.typ == tokenTypeEOF:
			return tree, nil
		}

		p.pos++
	}

	return tree, nil
}

func (p *parser) parseText() ([]interface{}, error) {
	if p.currentToken().typ != tokenTypeSeparatorOpen {
		return nil, errors.New(`text does not start with "{"`)
	}

	var text []interface{}

	for !p.isEOF() {
		token := p.nextToken()

		switch token.typ {
		case tokenTypeText:
			text = append(text, NodeText{Text: token.val})
		case tokenTypeSeparatorOpen:
			variable, err := p.parseVariable()
			if err != nil {
				return nil, fmt.Errorf("parse variable: %w", err)
			}

			text = append(text, variable)
			p.pos++
		case tokenTypeSeparatorClose:
			p.pos++

			return text, nil
		case tokenTypeKeyword, tokenTypeLiteral,
			tokenTypeFunction, tokenTypeVariable,
			tokenTypeEOF, tokenTypeError,
			tokenTypeOpeningFunction, tokenTypeClosingFunction:
		}
	}

	return nil, fmt.Errorf("invalid text node")
}

func (p *parser) parseMatch() (NodeMatch, error) {
	var match NodeMatch

	if v := p.currentToken(); v.typ != tokenTypeKeyword || v.val != KeywordMatch {
		return match, errors.New(`match node does not start with "match"`)
	}

	for !p.isEOF() {
		token := p.currentToken()

		if token.typ != tokenTypeKeyword {
			return NodeMatch{}, fmt.Errorf("invalid match")
		}

		switch token.val {
		case KeywordMatch:
			p.pos++

			expr, err := p.parseExpr()
			if err != nil {
				return NodeMatch{}, fmt.Errorf("parse expr: %w", err)
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
	if p.currentToken().typ != tokenTypeSeparatorOpen {
		return NodeExpr{}, errors.New(`expression does not start with "{"`)
	}

	var expr NodeExpr

	for !p.isEOF() {
		token := p.nextToken()

		switch token.typ {
		case tokenTypeVariable:
			expr.Value = NodeVariable{Name: token.val}
		case tokenTypeFunction:
			function, err := p.parseFunction()
			if err != nil {
				return NodeExpr{}, fmt.Errorf("parse function: %w", err)
			}

			expr.Function = function
		case tokenTypeSeparatorClose:
			p.pos++

			return expr, nil
		case tokenTypeKeyword, tokenTypeSeparatorOpen,
			tokenTypeLiteral, tokenTypeText,
			tokenTypeEOF, tokenTypeError,
			tokenTypeOpeningFunction, tokenTypeClosingFunction:
		}
	}

	return NodeExpr{}, fmt.Errorf("invalid expression node")
}

func (p *parser) parseVariant() (NodeVariant, error) {
	var variant NodeVariant

	literal := p.nextToken()

	if literal.typ != tokenTypeLiteral {
		return NodeVariant{}, fmt.Errorf(`literal does not follow keyword "when"`)
	}

	variant.Keys = append(variant.Keys, literal.val)

	p.pos++

	text, err := p.parseText()
	if err != nil {
		return NodeVariant{}, fmt.Errorf("parse text: %w", err)
	}

	variant.Message = append(variant.Message, text...)

	return variant, nil
}

func (p *parser) parseFunction() (NodeFunction, error) {
	var function NodeFunction

	if p.tokens[p.pos-1].typ != tokenTypeVariable {
		return function, errors.New(`function does not follow variable`)
	}

	function.Name = p.currentToken().val

	return function, nil
}

func (p *parser) parseVariable() (NodeVariable, error) {
	p.pos++

	var variable NodeVariable

	if p.tokens[p.pos-1].typ != tokenTypeSeparatorOpen {
		return variable, errors.New(`function does not follow placeholder open`)
	}

	variable.Name = p.currentToken().val

	return variable, nil
}

func (p *parser) currentToken() Token {
	return p.tokens[p.pos]
}

func (p *parser) nextToken() Token {
	p.pos++

	return p.tokens[p.pos]
}

func (p *parser) isEOF() bool { return p.tokens[p.pos].typ == tokenTypeEOF }
