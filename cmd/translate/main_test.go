package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	go main()

	os.Exit(m.Run())
}
