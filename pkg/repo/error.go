package repo

import (
	"errors"

	"golang.org/x/text/language"
)

var ErrNotFound = errors.New("entity not found")

type LoadMessagesOpts struct {
	FilterLanguages []language.Tag
}
