package messageformat

import "errors"

type AST []Node

// Node is a base node of Abstract Syntax Tree.
// Every node in AST implements Node interface.
type Node interface{}

type NodeMatch struct {
	Node

	Selectors []NodeExpr
	Variants  []NodeVariant
}

type NodeVariable struct {
	Node

	Name string
}

type NodeLiteral struct {
	Node

	Value string
}

type NodeText struct {
	Node

	Text string
}

type NodeExpr struct {
	Node

	Value    Node // NodeLiteral | NodeVariable
	Function NodeFunction
}

// isEmpty returns true if NodeExpr has zero-value.
func (n NodeExpr) isEmpty() bool {
	if n.Value == nil && n.Function.Options == nil && n.Function.Name == "" {
		return true
	}

	return false
}

type NodeVariant struct {
	Node

	Keys    []string
	Message []Node
}

type NodeFunction struct {
	Node

	Name    string
	Options []NodeOption
}

// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#options
type NodeOption struct {
	Node

	Value any // string | int | ...
	Name  string
}

type TextNodes []TextNode

// TextNode contains value and pointer to textNode in abstract syntax tree.
type TextNode struct {
	Ptr *Node
	NodeText
}

// GetTextNodes returns all textNodes found in ASTs.
func GetTextNodes(asts []AST) TextNodes {
	textNodes := make(TextNodes, 0, len(asts))

	for i := range asts {
		textNodes.fromAST(asts[i])
	}

	return textNodes
}

// fromAST retrieves textNodes from abstract syntax tree.
func (t *TextNodes) fromAST(ast AST) {
	for i := range ast {
		switch v := ast[i].(type) {
		case NodeText:
			*t = append(*t, TextNode{
				NodeText: v,
				Ptr:      &ast[i],
			})
		case NodeVariant:
			t.fromAST(v.Message)
		case NodeMatch:
			for i := range v.Variants {
				t.fromAST(v.Variants[i].Message)
			}
		}
	}
}

// GetTexts returns all texts from textNodes.
func (t TextNodes) GetTexts() []string {
	texts := make([]string, len(t))

	for i := range t {
		texts[i] = t[i].Text
	}

	return texts
}

// OverwriteTexts replaces texts stored in textNodes for all ASTs.
func (t TextNodes) OverwriteTexts(text []string) error {
	if len(t) != len(text) {
		return errors.New("text node count does not match number of text elements")
	}

	for i := range text {
		*t[i].Ptr = NodeText{Text: text[i]}
	}

	return nil
}
