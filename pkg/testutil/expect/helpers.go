package expect

import (
	"testing"
)

func Equal[T comparable](t *testing.T, want, got T) {
	t.Helper()

	if got != want {
		t.Errorf("want '%v', got '%v'", want, got)
	}
}
