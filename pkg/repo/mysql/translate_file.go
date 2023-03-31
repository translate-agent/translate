package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
)

// ---------------Messages Scanner/Valuer--------------.
type messages struct {
	model.Messages
}

func (m *messages) Scan(src any) error {
	data, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type '%T'", src)
	}

	var messages []model.Message

	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("unmarshal messages: %w", err)
	}

	m.Messages.Messages = messages

	return nil
}

func (m messages) Value() (driver.Value, error) {
	data, err := json.Marshal(m.Messages.Messages)
	if err != nil {
		return nil, fmt.Errorf("marshal messages: %w", err)
	}

	return data, nil
}

// ---------------Language Scanner/Valuer--------------

type languageTag struct {
	language.Tag
}

func (l *languageTag) Scan(src any) error {
	data, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("unsupported type '%T'", src)
	}

	str := string(data)

	var err error

	l.Tag, err = language.Parse(str)
	if err != nil {
		return fmt.Errorf("parse language: %w", err)
	}

	return nil
}

func (l languageTag) Value() (driver.Value, error) {
	return l.String(), nil
}

//--------------------Repo Implementation--------------------

func (r *Repo) SaveTranslateFile(
	ctx context.Context,
	serviceID uuid.UUID,
	translateFile *model.TranslateFile,
) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	if translateFile.ID == uuid.Nil {
		translateFile.ID = uuid.New()
	}

	var (
		lang = languageTag{Tag: translateFile.Language}
		msgs = messages{Messages: translateFile.Messages}
	)

	query := `INSERT INTO translate_file (
		id, service_id, 
		language, messages
		) 
		VALUES (
		UUID_TO_BIN(?), UUID_TO_BIN(?), 
		?, ?
		)
		ON DUPLICATE KEY UPDATE
    messages = VALUES(messages);`

	_, err = r.db.ExecContext(ctx, query,
		translateFile.ID, serviceID,
		lang, msgs,
	)

	if err != nil {
		return fmt.Errorf("repo: insert translate file: %w", err)
	}

	return nil
}

func (r *Repo) LoadTranslateFile(
	ctx context.Context,
	serviceID uuid.UUID,
	language language.Tag) (
	*model.TranslateFile, error,
) {
	query := `SELECT id, language, messages FROM translate_file WHERE service_id = UUID_TO_BIN(?) AND language = ?`

	row := r.db.QueryRowContext(ctx, query, serviceID, languageTag{Tag: language})

	var (
		id   uuid.UUID
		lang languageTag
		msgs messages
	)

	switch err := row.Scan(
		&id,
		&lang,
		&msgs,
	); {
	default:
		return &model.TranslateFile{ID: id, Language: lang.Tag, Messages: msgs.Messages}, nil
	case errors.Is(err, sql.ErrNoRows):
		return nil, repo.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("repo: select translate file: %w", err)
	}
}
