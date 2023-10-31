package messageformat

import (
	"errors"
)

type NodeMatch struct {
	Selectors []NodeExpr
	Variants  []NodeVariant
}

type NodeVariable struct {
	Name string
}

type NodeLiteral struct {
	Value string
}

type NodeText struct {
	Text string
}

type NodeExpr struct {
	Value    interface{} // NodeLiteral | NodeVariable
	Function NodeFunction
}

type NodeVariant struct {
	Keys    []string
	Message []interface{}
}

type NodeFunction struct {
	Name    string
	Options []NodeOption
}

type NodeOption struct {
	Value interface{}
	Name  string
}

// AST - abstract syntax tree.
type AST []interface{}

type TextNodes []TextNode

// TextNode contains value and pointer to textNode in abstract syntax tree.
type TextNode struct {
	Ptr *interface{}
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
		default:
			// noop
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

// GetText returns all text from textNodes.
func (t TextNodes) GetText() []string {
	texts := make([]string, len(t))

	for i := range t {
		texts[i] = t[i].Text
	}

	return texts
}

// OverwriteText replaces text stored in textNodes for all ASTs.
func (t TextNodes) OverwriteText(text []string) error {
	if len(t) != len(text) {
		return errors.New("text node count does not match number of text elements")
	}

	for i := range text {
		*t[i].Ptr = NodeText{Text: text[i]}
	}

	return nil
}
