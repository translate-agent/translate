package model

import "github.com/google/uuid"

type Service struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
}
