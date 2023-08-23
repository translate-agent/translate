package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"golang.org/x/exp/slices"
	"golang.org/x/text/language"
)

type Messages struct {
	Language language.Tag
	Messages []Message
	Original bool
}

type MessagesSlice []Messages

// Clone() returns a deep copy of MessagesSlice.
func (m MessagesSlice) Clone() MessagesSlice {
	if len(m) == 0 {
		return nil
	}

	msgs := make(MessagesSlice, len(m))

	copy(msgs, m)

	for i := range m {
		msgs[i].Messages = slices.Clone(m[i].Messages)
	}

	return msgs
}

// SplitOriginal returns a pointer to the original and other messages.
func (m MessagesSlice) SplitOriginal() (original *Messages, others MessagesSlice) {
	others = m

	for i := range m {
		if m[i].Original {
			original = &m[i]

			if len(m) > 2 { //nolint:gomnd
				others[i] = others[len(m)-1]
				others = others[:len(m)-1]
			}

			return
		}
	}

	return
}

type Message struct {
	ID          string
	PluralID    string
	Message     string
	Description string
	Positions   Positions
	Status      MessageStatus
}

type MessageStatus int32

const (
	MessageStatusUntranslated MessageStatus = iota
	MessageStatusFuzzy
	MessageStatusTranslated
)

func (s MessageStatus) String() string {
	switch s {
	default:
		return ""
	case MessageStatusUntranslated:
		return "UNTRANSLATED"
	case MessageStatusFuzzy:
		return "FUZZY"
	case MessageStatusTranslated:
		return "TRANSLATED"
	}
}

// Value implements driver.Valuer interface.
func (s MessageStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan implements sql.Scanner interface.
func (s *MessageStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	default:
		return fmt.Errorf("unknown type %+v, expected string", v)
	case []byte:
		switch string(v) {
		case MessageStatusUntranslated.String():
			*s = MessageStatusUntranslated
		case MessageStatusFuzzy.String():
			*s = MessageStatusFuzzy
		case MessageStatusTranslated.String():
			*s = MessageStatusTranslated
		default:
			return fmt.Errorf("unknown message status: %+v", v)
		}
	}

	return nil
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
