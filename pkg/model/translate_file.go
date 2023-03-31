package model

import (
	"github.com/google/uuid"
	"golang.org/x/text/language"
)

type TranslateFile struct {
	Language language.Tag
	Messages Messages
	ID       uuid.UUID
}
