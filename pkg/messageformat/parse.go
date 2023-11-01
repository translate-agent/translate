package messageformat

import (
	"errors"
	"fmt"
	"strings"
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

// Parse parses a MessageFormat2 string and returns Abstract Syntax Tree for it.
//
// Examples:
//
// Simple cases:
//
//	{"Hello World"} -> AST{NodeText{Text: "Hello World"}}
//	{"Hello {$user}"} -> AST{NodeText{Text: "Hello "}, NodeExpr{Value: NodeVariable{Name: "user"}}}
//
// Complex cases:
//
//	{"Hello {:Placeholder format=printf type=string}"} -> AST{NodeText{Text: "Hello "}, NodeExpr{Function: NodeFunction{Name: "Placeholder", Options: []NodeOption{{Name: "format", Value: "printf"}, {Name: "type", Value: "string"}}}}}
//
//	"match {$count} when 1 {One egg} when * {Multiple eggs}" -> AST{NodeMatch{Selectors: []NodeExpr{NodeExpr{Value: NodeVariable{Name: "count"}}}, Variants: []NodeVariant{NodeVariant{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "One egg"}}}, NodeVariant{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "Multiple eggs"}}}}}}
//
//nolint:lll
func Parse(text string) (AST, error) {
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

func (p *parser) parse() (AST, error) {
	var tree AST

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
			nodes, err := p.parseInsideCurly()
			if err != nil {
				return nil, fmt.Errorf("parse text: %w", err)
			}

			tree = append(tree, nodes...)
		case token.typ == tokenTypeEOF:
			return tree, nil
		}

		p.pos++
	}

	return tree, nil
}

// parseInsideCurly parses texts and expressions inside curly braces.
func (p *parser) parseInsideCurly() ([]interface{}, error) {
	if p.currentToken().typ != tokenTypeSeparatorOpen {
		return nil, errors.New("exp does not start with \"{\"")
	}

	var nodes []interface{}

	for !p.isEOF() {
		token := p.nextToken()

		switch token.typ {
		case tokenTypeText:
			nodes = append(nodes, NodeText{Text: token.val})
		case tokenTypeSeparatorOpen:
			expr, err := p.parseExpr()
			if err != nil {
				return nil, fmt.Errorf("parse variable: %w", err)
			}

			nodes = append(nodes, expr)
		case tokenTypeSeparatorClose:
			return nodes, nil
		case tokenTypeKeyword, tokenTypeLiteral,
			tokenTypeFunction, tokenTypeVariable,
			tokenTypeEOF, tokenTypeError,
			tokenTypeOpeningFunction, tokenTypeClosingFunction, tokenTypeOption:
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
			p.pos++ // skip "match" token

			expr, err := p.parseExpr()
			p.pos++ // skip "}" token

			if err != nil {
				return NodeMatch{}, fmt.Errorf("parse expr: %w", err)
			}

			match.Selectors = append(match.Selectors, expr)
		case KeywordWhen:
			variant, err := p.parseVariant()
			if err != nil {
				return NodeMatch{}, fmt.Errorf("parse variant: %w", err)
			}
			p.pos++ // skip "}" token

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
			expr.Function.Name = token.val
		case tokenTypeOption:
			expr.Function.Options = append(expr.Function.Options, p.parseOption())
		case tokenTypeSeparatorClose:
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

	nodes, err := p.parseInsideCurly()
	if err != nil {
		return NodeVariant{}, fmt.Errorf("parse text: %w", err)
	}

	variant.Message = append(variant.Message, nodes...)

	return variant, nil
}

func (p *parser) parseOption() NodeOption {
	split := strings.Split(p.currentToken().val, "=")

	return NodeOption{Name: split[0], Value: split[1]}
}

func (p *parser) currentToken() Token {
	return p.tokens[p.pos]
}

func (p *parser) nextToken() Token {
	p.pos++

	return p.tokens[p.pos]
}

func (p *parser) isEOF() bool { return p.tokens[p.pos].typ == tokenTypeEOF }
