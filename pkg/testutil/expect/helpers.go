package expect

import (
	"strings"
	"testing"

	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func NoError(t *testing.T, got error) {
	t.Helper()

	if got != nil {
		t.Errorf("want no error, got %s", got)
	}
}

func ErrorContains(t *testing.T, err error, str string) {
	t.Helper()

	if !strings.Contains(err.Error(), str) {
		t.Errorf("want error to contain '%s', got '%s'", str, err)
	}
}

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
