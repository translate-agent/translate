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

type dbMessageSlice []model.Message

type dbLanguageTag struct {
	language.Tag
}

type dbTranslationFile struct {
	lang     dbLanguageTag
	messages dbMessageSlice
	id       uuid.UUID
}

// fromTranslationFile converts model.TranslationFile to dbTranslationFile.
func fromTranslationFile(translationFile *model.TranslationFile) *dbTranslationFile {
	return &dbTranslationFile{
		lang:     dbLanguageTag{Tag: translationFile.Messages.Language},
		messages: translationFile.Messages.Messages,
		id:       translationFile.ID,
	}
}

// toTranslationFile converts dbTranslationFile to model.TranslationFile.
func toTranslationFile(translationFile *dbTranslationFile) *model.TranslationFile {
	return &model.TranslationFile{
		ID: translationFile.id,
		Messages: model.Messages{
			Language: translationFile.lang.Tag,
			Messages: translationFile.messages,
		},
	}
}

// ---------------Messages Scanner/Valuer--------------.

func (d *dbMessageSlice) Scan(src any) error {
	data, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type '%T'", src)
	}

	var messages dbMessageSlice

	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("unmarshal messages: %w", err)
	}

	*d = messages

	return nil
}

func (d dbMessageSlice) Value() (driver.Value, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("marshal messages: %w", err)
	}

	return data, nil
}

// ---------------Language Scanner/Valuer--------------

func (d *dbLanguageTag) Scan(src any) error {
	data, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("unsupported type '%T'", src)
	}

	str := string(data)

	var err error

	d.Tag, err = language.Parse(str)
	if err != nil {
		return fmt.Errorf("parse language: %w", err)
	}

	return nil
}

func (d dbLanguageTag) Value() (driver.Value, error) {
	return d.Tag.String(), nil
}

//--------------------Repo Implementation--------------------

func (r *Repo) SaveTranslationFile(
	ctx context.Context,
	serviceID uuid.UUID,
	translationFile *model.TranslationFile,
) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	if translationFile.ID == uuid.Nil {
		translationFile.ID = uuid.New()
	}

	dbFile := fromTranslationFile(translationFile)

	query := `INSERT INTO translation_file (
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
		dbFile.id, serviceID,
		dbFile.lang, dbFile.messages,
	)

	if err != nil {
		return fmt.Errorf("repo: insert translate file: %w", err)
	}

	return nil
}

func (r *Repo) LoadTranslationFile(
	ctx context.Context,
	serviceID uuid.UUID,
	language language.Tag) (
	*model.TranslationFile, error,
) {
	query := `SELECT id, language, messages FROM translation_file WHERE service_id = UUID_TO_BIN(?) AND language = ?`

	row := r.db.QueryRowContext(ctx, query, serviceID, dbLanguageTag{Tag: language})

	var dbFile dbTranslationFile

	switch err := row.Scan(
		&dbFile.id,
		&dbFile.lang,
		&dbFile.messages,
	); {
	default:
		return toTranslationFile(&dbFile), nil
	case errors.Is(err, sql.ErrNoRows):
		return nil, repo.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("repo: select translate file: %w", err)
	}
}
