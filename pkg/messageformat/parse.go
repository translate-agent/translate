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
		case tokenTypeVariable, tokenTypeFunction, tokenTypeExpressionOpen,
			tokenTypeExpressionClose, tokenTypeText, tokenTypeOpeningFunction,
			tokenTypeClosingFunction, tokenTypeKeyword, tokenTypeLiteral,
			tokenTypeComplexMessageClose, tokenTypeComplexMessageOpen:
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

		// default complex msg open
		if p.pos == 0 && p.tokens[0].typ == tokenTypeComplexMessageOpen {
			p.pos++
			continue
		}

		// default complex msg close
		if p.pos == len(p.tokens)-2 && p.tokens[len(p.tokens)-2].typ == tokenTypeComplexMessageClose {
			p.pos++
			continue
		}

		switch {
		default:
			return nil, fmt.Errorf("unknown token: %+v", token)
		case token.typ == tokenTypeKeyword && token.val == KeywordMatch:
			match, err := p.parseMatch()
			if err != nil {
				return nil, fmt.Errorf("parse match: %w", err)
			}

			tree = append(tree, match)
		case token.typ == tokenTypeComplexMessageOpen:
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
	if p.currentToken().typ != tokenTypeComplexMessageOpen {
		return nil, errors.New(`text does not start with "{"`)
	}

	var text []interface{}

	for !p.isEOF() {
		token := p.nextToken()

		switch token.typ {
		case tokenTypeText:
			text = append(text, NodeText{Text: token.val})
		case tokenTypeExpressionOpen:
			variable, err := p.parseVariable()
			if err != nil {
				return nil, fmt.Errorf("parse variable: %w", err)
			}

			text = append(text, variable)
			p.pos++
		case tokenTypeComplexMessageClose:
			return text, nil
		case tokenTypeKeyword, tokenTypeLiteral,
			tokenTypeFunction, tokenTypeVariable,
			tokenTypeEOF, tokenTypeError,
			tokenTypeOpeningFunction, tokenTypeClosingFunction,
			tokenTypeComplexMessageOpen, tokenTypeExpressionClose:
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
	if p.currentToken().typ != tokenTypeExpressionOpen {
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
		case tokenTypeExpressionClose:
			p.pos++

			return expr, nil
		case tokenTypeKeyword, tokenTypeExpressionOpen,
			tokenTypeLiteral, tokenTypeText,
			tokenTypeEOF, tokenTypeError,
			tokenTypeOpeningFunction, tokenTypeClosingFunction,
			tokenTypeComplexMessageClose, tokenTypeComplexMessageOpen:
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

	// shift pos
	if p.nextToken().typ == tokenTypeComplexMessageClose && p.pos == len(p.tokens)-2 {
		p.pos++
	}

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

	if p.tokens[p.pos-1].typ != tokenTypeExpressionOpen {
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
