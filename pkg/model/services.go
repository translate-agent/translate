package model

import "github.com/google/uuid"

type Service struct {
	Name string `protoName:"name"`
	ID   uuid.UUID
}
