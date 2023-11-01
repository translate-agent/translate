package messageformat

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// keywords.
const (
	match = "match"
	when  = "when"
)

// message stores MF2 message.
// Specification draft for MF2 syntax:
// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md
type message struct {
	strings.Builder
}

// Compile traverses abstract syntax tree
// constructing message from nodes, returns message string.
func Compile(ast AST) (string, error) {
	if len(ast) == 0 {
		return "", errors.New("AST must contain at least one node")
	}

	var m message

	if err := m.fromAST(ast); err != nil {
		return "", fmt.Errorf("message from abstract syntax tree: %w", err)
	}

	return m.String(), nil
}

// writeExpr writes MF2 expression to the message.
func (m *message) writeExpr(n NodeExpr) error {
	if reflect.DeepEqual(n, NodeExpr{}) {
		return errors.New("expression node must not be empty")
	}

	var expr message

	switch v := n.Value.(type) {
	default:
		return fmt.Errorf("unsupported node type '%T' for expression value", n.Value)
	case nil:
		break
	case NodeVariable:
		expr.WriteString("$" + v.Name)

		if n.Function.Name != "" {
			expr.WriteByte(' ')
		}
	}

	if err := expr.writeFunc(n.Function); err != nil {
		return fmt.Errorf("write function: %w", err)
	}

	m.WriteString("{" + expr.String() + "}")

	return nil
}

// writeMatch writes MF2 matcher body to the message.
func (m *message) writeMatch(n NodeMatch) error {
	if len(n.Selectors) == 0 {
		return errors.New("there must be at least one selector")
	}

	m.WriteString(match + " ")

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
			return fmt.Errorf("number of keys %d for variant #%d don't match number of match selectors %d",
				len(n.Variants[i].Keys), i, numberOfSelectors)
		}

		if err := m.writeVariant(n.Variants[i]); err != nil {
			return fmt.Errorf("write variant: %w", err)
		}
	}

	return nil
}

// writeVariant writes MF2 matcher variant to the message.
func (m *message) writeVariant(n NodeVariant) error {
	for i, v := range n.Keys {
		if i == 0 {
			m.WriteString(" " + when + " ")
		}

		m.WriteString(v + " ")
	}

	if err := m.fromAST(n.Message); err != nil {
		return fmt.Errorf("from AST: %w", err)
	}

	return nil
}

// writeFunc writes MF2 expression function to the message.
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

// writeText writes MF2 text to the message.
func (m *message) writeText(n NodeText, pos int, ast []interface{}) {
	switch {
	default: // nodeText set in middle
		m.WriteString(n.Text)
	case len(ast) == 1: // nodeText is only element
		m.WriteString("{" + n.Text + "}")
	case pos == 0: // nodeText is first element
		m.WriteString("{" + n.Text)
	case pos == len(ast)-1: // nodeText is last element
		m.WriteString(n.Text + "}")
	}
}

// writeVar writes MF2 variable to the message.
func (m *message) writeVar(n NodeVariable, ast []interface{}) {
	switch len(ast) {
	default:
		m.WriteString("{$" + n.Name + "}")
	case 1:
		m.WriteString("$" + n.Name)
	}
}

// fromAST constructs message from nodes in abstract syntax tree.
func (m *message) fromAST(ast []interface{}) error {
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
			m.writeVar(v, ast)
		case NodeText:
			m.writeText(v, i, ast)
		}
	}

	return nil
}
