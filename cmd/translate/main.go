package main

import (
	"go.expect.digital/translate/cmd/translate/service"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func main() {
	service.Serve()
}
