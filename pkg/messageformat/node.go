package messageformat

type AST []interface{}

func (n *AST) Append(node ...interface{}) {
	*n = append(*n, node...)
}

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

// https://github.com/unicode-org/message-format-wg/blob/main/spec/syntax.md#options
type NodeOption struct {
	Value interface{}
	Name  string
}
