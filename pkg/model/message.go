package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"golang.org/x/text/language"
)

type Messages struct {
	Language language.Tag
	Messages []Message
	Original bool
}

type Message struct {
	ID          string
	PluralID    string
	Message     string
	Description string
	Positions   Positions
	Fuzzy       bool
}

type Positions []string

// Value implements driver.Valuer interface.
func (p Positions) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("json marshal positions: %w", err)
	}

	return b, nil
}

// Scan implements sql.Scanner interface.
func (p *Positions) Scan(value interface{}) error {
	switch v := value.(type) {
	default:
		return fmt.Errorf("unknown type %+v, expected []byte", v)
	case nil:
		*p = nil
		return nil
	case []byte:
		if err := json.Unmarshal(v, &p); err != nil {
			return fmt.Errorf("json unmarshal positions: %w", err)
		}

		return nil
	}
}
