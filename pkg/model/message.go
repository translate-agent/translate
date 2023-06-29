package model

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

type Messages struct {
	Language language.Tag
	Messages []Message
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
	switch {
	case p == nil:
		return nil, nil
	case len(p) == 0:
		return "", nil
	default:
		return strings.Join(p, ", "), nil
	}
}

// Scan implements sql.Scanner interface.
func (p *Positions) Scan(value interface{}) error {
	var positions sql.NullString
	if err := positions.Scan(value); err != nil {
		return fmt.Errorf("scan positions: %w", err)
	}

	if positions.Valid {
		if len(positions.String) > 0 {
			*p = strings.Split(positions.String, ", ")
		} else {
			*p = Positions{}
		}
	}

	return nil
}
