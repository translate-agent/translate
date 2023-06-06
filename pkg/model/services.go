package model

import "github.com/google/uuid"

type Service struct {
	Name string
	ID   uuid.UUID
}
