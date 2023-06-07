package main

import (
	"context"
	"log"

	"go.expect.digital/translate/cmd/client/cmd"
)

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		log.Panic(err)
	}
}
