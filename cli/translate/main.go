package main

import (
	"os"

	"go.expect.digital/translate/cli/translate/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
