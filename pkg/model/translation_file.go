package model

import (
	"github.com/google/uuid"
)

type TranslationFile struct {
	Messages Messages
	ID       uuid.UUID
}
