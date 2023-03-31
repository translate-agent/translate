package model

import (
	"github.com/google/uuid"
)

type TranslateFile struct {
	Messages Messages
	ID       uuid.UUID
}
