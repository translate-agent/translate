package main

import (
	"log"

	"go.expect.digital/translate/cmd/client/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Panic(err)
	}
}
