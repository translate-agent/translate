package messageformat

import (
	"fmt"
	"strings"
)

type parser struct {
	lex    *lexer
	text   string
	tokens []Token
	pos    int
}

// collect collects all tokens from the lexer and stores them in the parser.
func (p *parser) collect() error {
	for v := p.lex.nextToken(); ; v = p.lex.nextToken() {
		if v.typ == tokenTypeError {
			return fmt.Errorf("lex: %s", v.val)
		}

		p.tokens = append(p.tokens, v)

		if v.typ == tokenTypeEOF {
			return nil
		}
	}
}

// Parse parses the given text into an MessageFormat2 AST (abstract syntax tree).
func Parse(input string) (AST, error) {
	p := parser{lex: lex(input), text: input}

	if err := p.collect(); err != nil {
		return nil, fmt.Errorf("collect tokens: %w", err)
	}

	return p.parse(), nil
}

// parse parses the tokens collected by the lexer.
func (p *parser) parse() AST {
	var tree AST

	for token := p.currentToken(); !p.isEOF(); token = p.nextToken() {
		parseNode := func() interface{} { return nil }

		switch token.typ {
		case tokenTypeText:
			parseNode = func() interface{} { return p.parseText2() }
		case tokenTypeKeyword:
			parseNode = func() interface{} { return p.parseMatch2() }
		case tokenTypeFunction:
			parseNode = func() interface{} { return p.parseExpression2() }
		case tokenTypeVariable:
			parseNode = func() interface{} { return p.parseVariable2() }
		}

		if node := parseNode(); node != nil {
			tree.Append(node)
		}
	}

	return tree
}

func (p *parser) currentToken() Token {
	return p.tokens[p.pos]
}

func (p *parser) nextToken() Token {
	p.pos++

	return p.tokens[p.pos]
}

func (p *parser) peek() Token {
	return p.tokens[p.pos+1]
}

func (p *parser) isEOF() bool { return p.tokens[p.pos].typ == tokenTypeEOF }

// parsers

func (p *parser) parseText2() NodeText {
	return NodeText{p.currentToken().val}
}

func (p *parser) parseMatch2() NodeMatch {
	var (
		match     NodeMatch
		whenCount int
	)

	// count when keywords
	for _, v := range p.tokens {
		if v.typ == tokenTypeKeyword && v.val == KeywordWhen {
			whenCount++
		}
	}

	var count int

	for token := p.currentToken(); ; token = p.nextToken() {
		switch token.typ {
		case tokenTypeVariable, tokenTypeLiteral, tokenTypeFunction:
			match.Selectors = append(match.Selectors, p.parseExpression2())
		case tokenTypeKeyword:
			if token.val != KeywordWhen {
				continue
			}
			count++

			match.Variants = append(match.Variants, p.parseVariant2())

			if count == whenCount {
				return match
			}
		}
	}
}

func (p *parser) parseVariant2() NodeVariant {
	var variant NodeVariant

	for token := p.currentToken(); ; token = p.nextToken() {
		switch token.typ {
		case tokenTypeLiteral:
			variant.Keys = append(variant.Keys, token.val)
		case tokenTypeText:
			variant.Message = append(variant.Message, p.parseText2())
		case tokenTypeVariable:
			if p.peek().typ == tokenTypeSeparatorClose {
				variant.Message = append(variant.Message, p.parseVariable2())
				p.nextToken()
			} else {
				variant.Message = append(variant.Message, p.parseExpression2())
			}

		case tokenTypeSeparatorClose:
			return variant
		}
	}
}

func (p *parser) parseVariable2() NodeVariable {
	return NodeVariable{Name: p.currentToken().val}
}

func (p *parser) parseLiteral2() NodeLiteral {
	return NodeLiteral{Value: p.currentToken().val}
}

func (p *parser) parseExpression2() NodeExpr {
	var expr NodeExpr

	for token := p.currentToken(); ; token = p.nextToken() {
		switch token.typ {
		case tokenTypeVariable:
			expr.Value = p.parseVariable2()
		case tokenTypeLiteral:
			expr.Value = p.parseLiteral2()
		case tokenTypeFunction:
			expr.Function = p.parseFunction2()
			return expr
		case tokenTypeSeparatorClose:
			return expr
		}
	}
}

func (p *parser) parseOption2() NodeOption {
	// key=value
	splitted := strings.Split(p.currentToken().val, "=")

	return NodeOption{Name: splitted[0], Value: splitted[1]}
}

func (p *parser) parseFunction2() NodeFunction {
	var function NodeFunction

	for token := p.currentToken(); ; token = p.nextToken() {
		switch token.typ {
		case tokenTypeOption:
			function.Options = append(function.Options, p.parseOption2())
		case tokenTypeFunction:
			function.Name = token.val
		case tokenTypeSeparatorClose:
			return function
		}
	}
}
