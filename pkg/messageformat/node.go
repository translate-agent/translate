package messageformat

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
	Name string
}
