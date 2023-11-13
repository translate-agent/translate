package model

import "github.com/google/uuid"

type Service struct {
	Name string    `json:"name" protoName:"name"`
	ID   uuid.UUID `json:"id"`
}
