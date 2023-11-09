package messageformat

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
