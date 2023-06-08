package googletranslate

import (
	"context"
	"io"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

type GoogleClient interface {
	Translate(ctx context.Context, inputs []string, target language.Tag, opts *translate.Options) ([]translate.Translation, error) //nolint:lll
	io.Closer
}

type GoogleTranslate struct {
	client GoogleClient
}

func NewGoogleTranslate(c GoogleClient) *GoogleTranslate {
	return &GoogleTranslate{client: c}
}
