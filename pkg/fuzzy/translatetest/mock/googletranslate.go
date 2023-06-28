package mock

import (
	"context"
	"errors"

	"cloud.google.com/go/translate"
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/language"
)

// MockGoogleTranslateClient is a mock implementation of the Google Translate client.
type MockGoogleTranslateClient struct{}

// Translate mocks the Translate method of the Google Translate client.
func (m *MockGoogleTranslateClient) Translate(
	_ context.Context,
	inputs []string,
	target language.Tag,
	_ *translate.Options,
) ([]translate.Translation, error) {
	// Mock the Bad request error for unsupported language.Afrikaans.
	if target == language.Afrikaans {
		return nil, errors.New("mock: bad request: unsupported language")
	}

	translations := make([]translate.Translation, 0, len(inputs))

	for range inputs {
		translations = append(translations, translate.Translation{Text: gofakeit.SentenceSimple()})
	}

	return translations, nil
}

func (m *MockGoogleTranslateClient) Close() error { return nil }
