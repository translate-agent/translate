package model

import (
	"database/sql/driver"
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
