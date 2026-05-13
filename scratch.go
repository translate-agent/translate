package main

import (
	"fmt"

	"go.expect.digital/mf2/builder"
	"go.expect.digital/mf2/parse"
)

func main() {
	b := builder.NewBuilder()
	b.Input(builder.Var("count").Func("number"))
	b.Match(parse.Variable("count"))
	b.Keys(1).Text("apple")
	b.Keys("*").Text("apples")
	out, err := b.Build()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(out)
}
