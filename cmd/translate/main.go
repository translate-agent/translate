package main

import (
	"go.expect.digital/translate/cmd/translate/service"
	_ "go.uber.org/automaxprocs" // Automatically set GOMAXPROCS to match Linux container CPU quota
)

func main() {
	service.Serve()
}
