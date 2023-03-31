package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"golang.org/x/text/language"
)

type Language struct {
	language.Tag
}

type Schema struct {
	translatev1.Schema
}

type TranslateFile struct {
	Language Language
	Messages Messages
	ID       uuid.UUID
}

// ---------------Messages Scanner/Valuer--------------

func (m *Messages) Scan(src any) error {
	data, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type '%T'", src)
	}

	var messages []Message

	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("unmarshal messages: %w", err)
	}

	m.Messages = messages

	return nil
}

func (m Messages) Value() (driver.Value, error) {
	data, err := json.Marshal(m.Messages)
	if err != nil {
		return nil, fmt.Errorf("marshal messages: %w", err)
	}

	return data, nil
}

// ---------------Language Scanner/Valuer--------------

func (l *Language) Scan(src any) error {
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []uint8:
		str = string(v)
	default:
		return fmt.Errorf("unsupported type '%T'", src)
	}

	var err error

	l.Tag, err = language.Parse(str)
	if err != nil {
		return fmt.Errorf("parse language: %w", err)
	}

	return nil
}

func (l Language) Value() (driver.Value, error) {
	return l.String(), nil
}
