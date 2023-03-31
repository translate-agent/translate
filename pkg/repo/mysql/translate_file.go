package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
)

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

	query := `INSERT INTO translate_file (
		id, service_id, language, messages
		) 
		VALUES (
		UUID_TO_BIN(?), UUID_TO_BIN(?), 
		?, ?
		)
		ON DUPLICATE KEY UPDATE
    messages = VALUES(messages);`

	_, err = r.db.ExecContext(ctx, query,
		translateFile.ID, serviceID,
		translateFile.Language, translateFile.Messages,
	)

	if err != nil {
		return fmt.Errorf("repo: insert translate file: %w", err)
	}

	return nil
}

func (r *Repo) LoadTranslateFile(
	ctx context.Context,
	serviceID uuid.UUID,
	language model.Language) (
	*model.TranslateFile, error,
) {
	query := `SELECT id, language, messages FROM translate_file WHERE service_id = UUID_TO_BIN(?) AND language = ?`

	row := r.db.QueryRowContext(ctx, query, serviceID, language)

	var translateFile model.TranslateFile

	switch err := row.Scan(
		&translateFile.ID,
		&translateFile.Language,
		&translateFile.Messages,
	); {
	default:
		return &translateFile, nil
	case errors.Is(err, sql.ErrNoRows):
		return nil, repo.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("repo: select translate file: %w", err)
	}
}
