package messageformat

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// message stores MF2 message.
// Specification draft for MF2 syntax:
// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md
type message struct {
	bytes.Buffer
}

/*
MarshalText encodes abstract syntax tree into UTF-8-encoded 'Message Format v2' text and returns the result.
MarshalText implements the encoding.MarshalText interface for custom AST type.

Example:

Input:

	AST{
			NodeMatch{
				Selectors: []NodeExpr{
					{
						Value:    NodeVariable{Name: "count"},
						Function: NodeFunction{Name: "number"},
					},
				},
				Variants: []NodeVariant{
					{
						Keys:    []string{"*"},
						Message: []interface{}{NodeText{Text: "Hello, world\\!"}},
					},
				},
			},
		}

Output:

	[]byte("match {$count :number} when * {Hello, world\\!}"), nil
*/
func (a AST) MarshalText() ([]byte, error) {
	var m message

	if err := m.fromAST(a); err != nil {
		return nil, fmt.Errorf("message from abstract syntax tree: %w", err)
	}

	return m.Bytes(), nil
}

/*
UnmarshalText decodes UTF-8-encoded 'Message Format v2' text to receiving AST.
UnmarshalText implements the encoding.UnmarshalText interface for custom AST type.

Example:

Input:

	[]byte("match {$count :number} when * {Hello, world\\!}"), nil

Output:

	AST{
			NodeMatch{
				Selectors: []NodeExpr{
					{
						Value:    NodeVariable{Name: "count"},
						Function: NodeFunction{Name: "number"},
					},
				},
				Variants: []NodeVariant{
					{
						Keys:    []string{"*"},
						Message: []interface{}{NodeText{Text: "Hello, world\\!"}},
					},
				},
			},
		}
*/
func (a *AST) UnmarshalText(text []byte) error {
	var err error

	if *a, err = Parse(string(text)); err != nil {
		return fmt.Errorf("parse MF2 message: %w", err)
	}

	return nil
}

// writeExpr writes expression from NodeExpr to the receiving message.
func (m *message) writeExpr(n NodeExpr) error {
	if n.isEmpty() {
		return errors.New("expression node must not be empty")
	}

	m.WriteByte('{')

	switch v := n.Value.(type) {
	default:
		return fmt.Errorf("unsupported node type '%T' for expression value", n.Value)
	case nil:
		break
	case NodeVariable:
		m.WriteString("$" + v.Name)

		if n.Function.Name != "" {
			m.WriteByte(' ')
		}
	}

	if err := m.writeFunc(n.Function); err != nil {
		return fmt.Errorf("write function: %w", err)
	}

	m.WriteByte('}')

	return nil
}

// writeMatch writes matcher from NodeMatch to the receiving message.
func (m *message) writeMatch(n NodeMatch) error {
	if len(n.Selectors) == 0 {
		return errors.New("there must be at least one selector")
	}

	m.WriteString(KeywordMatch + " ")

	for i := range n.Selectors {
		if i > 0 {
			m.WriteByte(' ')
		}

		if err := m.writeExpr(n.Selectors[i]); err != nil {
			return fmt.Errorf("write expression: %w", err)
		}
	}

	numberOfSelectors := len(n.Selectors)

	for i := range n.Variants {
		if len(n.Variants[i].Keys) != numberOfSelectors {
			return fmt.Errorf("number of keys '%d' for variant #%d don't match number of match selectors '%d'",
				len(n.Variants[i].Keys), i, numberOfSelectors)
		}

		if err := m.writeVariant(n.Variants[i]); err != nil {
			return fmt.Errorf("write variant: %w", err)
		}
	}

	return nil
}

// writeVariant writes match variant from NodeVariant to the receiving message.
func (m *message) writeVariant(n NodeVariant) error {
	for i, v := range n.Keys {
		if i == 0 {
			m.WriteString(" " + KeywordWhen + " ")
		}

		m.WriteString(v + " ")
	}

	if err := m.fromAST(n.Message); err != nil {
		return fmt.Errorf("from AST: %w", err)
	}

	return nil
}

// writeFunc writes function from NodeFunc to the receiving message.
// TODO: add ability to process markup-like functions e.g.
// {+button}Submit{-button} or {+link}cancel{-link}.
func (m *message) writeFunc(n NodeFunction) error {
	if n.Name == "" {
		return nil
	}

	m.WriteString(":" + n.Name)

	for i := range n.Options {
		m.WriteString(" " + n.Options[i].Name + "=")

		switch v := n.Options[i].Value.(type) {
		default:
			return fmt.Errorf("unsupported type '%T' for function option '%s':", n.Options[i].Value, n.Options[i].Name)
		case string:
			m.WriteString(v)
		case int:
			m.WriteString(strconv.Itoa(v))
		}
	}

	return nil
}

// fromAST traverses 'Message Format v2' nodes in the abstract syntax tree (AST)
// writes constructed message parts to the receiving message.
// Implementation follows MF2 design draft:
// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md
func (m *message) fromAST(ast AST) error {
	for i := range ast {
		switch v := ast[i].(type) {
		default:
			return fmt.Errorf("unsupported node type '%T'", ast[i])
		case NodeMatch:
			if err := m.writeMatch(v); err != nil {
				return fmt.Errorf("write match: %w", err)
			}
		case NodeExpr:
			if err := m.writeExpr(v); err != nil {
				return fmt.Errorf("write expression: %w", err)
			}
		case NodeVariable:
			switch len(ast) {
			default:
				m.WriteString("{$" + v.Name + "}")
			case 1:
				m.WriteString("$" + v.Name)
			}
		case NodeText:
			switch {
			default: // nodeText set in middle
				m.WriteString(v.Text)
			case len(ast) == 1: // nodeText is only element
				m.WriteString("{" + v.Text + "}")
			case i == 0: // nodeText is first element
				m.WriteString("{" + v.Text)
			case i == len(ast)-1: // nodeText is last element
				m.WriteString(v.Text + "}")
			}
		}
	}

	return nil
}
