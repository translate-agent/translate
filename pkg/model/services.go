package model

import "uuid"

type Service struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
}
