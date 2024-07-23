package expect

import (
	"testing"

	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func Service(t *testing.T, service *translatev1.Service) {
	t.Helper()

	if service == nil {
		t.Error("want service, got nil")
	}
}

func Equal[T comparable](t *testing.T, want, got T) {
	t.Helper()

	if got != want {
		t.Errorf("want '%v', got '%v'", want, got)
	}
}
