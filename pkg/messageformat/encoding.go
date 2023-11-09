package messageformat

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

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
	var buf bytes.Buffer

	if err := marshal(&buf, a); err != nil {
		return nil, fmt.Errorf("encode message format v2: %w", err)
	}

	return buf.Bytes(), nil
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

func writeExpr(buf *bytes.Buffer, n NodeExpr) error {
	if n.isEmpty() {
		return errors.New("expression node must not be empty")
	}

	buf.WriteByte('{')

	switch v := n.Value.(type) {
	default:
		return fmt.Errorf("unsupported node type '%T' for expression value", n.Value)
	case nil:
		break
	case NodeVariable:
		buf.WriteString("$" + v.Name)

		if n.Function.Name != "" {
			buf.WriteByte(' ')
		}
	}

	if err := writeFunc(buf, n.Function); err != nil {
		return fmt.Errorf("write function: %w", err)
	}

	buf.WriteByte('}')

	return nil
}

func writeMatch(buf *bytes.Buffer, n NodeMatch) error {
	if len(n.Selectors) == 0 {
		return errors.New("there must be at least one selector")
	}

	buf.WriteString(KeywordMatch + " ")

	for i := range n.Selectors {
		if i > 0 {
			buf.WriteByte(' ')
		}

		if err := writeExpr(buf, n.Selectors[i]); err != nil {
			return fmt.Errorf("write expression: %w", err)
		}
	}

	numberOfSelectors := len(n.Selectors)

	for i := range n.Variants {
		if len(n.Variants[i].Keys) != numberOfSelectors {
			return fmt.Errorf("number of keys '%d' for variant #%d don't match number of match selectors '%d'",
				len(n.Variants[i].Keys), i, numberOfSelectors)
		}

		if err := writeVariant(buf, n.Variants[i]); err != nil {
			return fmt.Errorf("write variant: %w", err)
		}
	}

	return nil
}

func writeVariant(buf *bytes.Buffer, n NodeVariant) error {
	for i, v := range n.Keys {
		if i == 0 {
			buf.WriteString(" " + KeywordWhen + " ")
		}

		buf.WriteString(v + " ")
	}

	if err := marshal(buf, n.Message); err != nil {
		return fmt.Errorf("encode message format v2: %w", err)
	}

	return nil
}

// TODO: add ability to process markup-like functions e.g.
// {+button}Submit{-button} or {+link}cancel{-link}.
func writeFunc(buf *bytes.Buffer, n NodeFunction) error {
	if n.Name == "" {
		return nil
	}

	buf.WriteString(":" + n.Name)

	for i := range n.Options {
		buf.WriteString(" " + n.Options[i].Name + "=")

		switch v := n.Options[i].Value.(type) {
		default:
			return fmt.Errorf("unsupported type '%T' for function option '%s'", n.Options[i].Value, n.Options[i].Name)
		case string:
			buf.WriteString(v)
		case int:
			buf.WriteString(strconv.Itoa(v))
		}
	}

	return nil
}

// marshal serializes AST nodes to text.
// Implementation follows MF2 design draft:
// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md
func marshal(buf *bytes.Buffer, ast AST) error {
	for i := range ast {
		switch v := ast[i].(type) {
		default:
			return fmt.Errorf("unsupported node type '%T'", ast[i])
		case NodeMatch:
			if err := writeMatch(buf, v); err != nil {
				return fmt.Errorf("write match: %w", err)
			}
		case NodeExpr:
			if err := writeExpr(buf, v); err != nil {
				return fmt.Errorf("write expression: %w", err)
			}
		case NodeVariable:
			switch len(ast) {
			default:
				buf.WriteString("{$" + v.Name + "}")
			case 1:
				buf.WriteString("$" + v.Name)
			}
		case NodeText:
			switch {
			default: // nodeText set in middle
				buf.WriteString(v.Text)
			case len(ast) == 1: // nodeText is only element
				buf.WriteString("{" + v.Text + "}")
			case i == 0: // nodeText is first element
				buf.WriteString("{" + v.Text)
			case i == len(ast)-1: // nodeText is last element
				buf.WriteString(v.Text + "}")
			}
		}
	}

	return nil
}
