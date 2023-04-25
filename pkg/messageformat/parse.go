package messageformat

import "fmt"

//match {$count :number} when 0 {No notifications} when 1 {You have one notification.} when * {You have {$count} notifications.}

//Tree {
//	root = NodeMatch{
//		Selectors: []NodeExpr{
//		{Value: NodeVariable{
//		Name: "count"},
//		Function: NodeFunction{
//		Name: "number"}}
//		},
//		Variants: []NodeVariant{
//		{Keys: []string{"1"}, Message: []interface{}{NodeText{Text: "You have one notification"}}},
//		{Keys: []string{"*"}, Message: []interface{}{NodeText{Text: "You have "}, NodeVariable{Name: "count"}, NodeText{Text: " notifications"}}},
//		},
//		},
//	lex = lexer{}
//	text = input
//}

type Tree struct {
	lex    *lexer
	text   string
	root   *NodeMatch
	tokens []token
}

type NodeMatch struct {
	Selectors []NodeExpr
	Variants  []NodeVariant
	lex       *lexer
	text      string
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

//type Tree struct {
//
//}

func (t *Tree) startParse(lex *lexer) {
	t.lex = lex
}

func Parse(text string) (*NodeMatch, error) {
	t := New(text)
	_, err := t.Parse(text)
	return t.root, err
}

func New(text string) *Tree {
	return &Tree{
		text: text,
	}
}

func (t *Tree) Parse(text string) (tree *Tree, err error) {
	//defer n.recover(&err)
	lexer := lex(text)
	t.startParse(lexer)
	t.collect()
	t.parse()
	return t, nil
}

func (t *Tree) collect() {
	for {
		v := t.lex.nextToken()
		fmt.Println(v)

		switch v.typ {
		default:
			t.tokens = append(t.tokens, v)
		case tokenTypeEOF, tokenTypeError:
			return
		case tokenTypeSeparatorOpen, tokenTypeSeparatorClose, tokenTypeKeyword:
			continue
		}
	}
}

// parse is the top-level parser for a template, essentially the same
// as itemList except it also parses {{define}} actions.
// It runs to EOF.
func (t *Tree) parse() {
	fmt.Println(t.tokens)
}
