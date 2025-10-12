package main

import (
	"context"
	"log"

	"go.expect.digital/translate/cmd/client/cmd"
)

func main() {
	err := cmd.Execute(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
