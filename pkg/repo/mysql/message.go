package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
)

func (r *Repo) SaveMessages(ctx context.Context, serviceID uuid.UUID, messages *model.Messages) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return &repo.DefaultError{Entity: "message", Err: err, Operation: "Start Transaction"}
	}

	defer tx.Rollback() //nolint:errcheck

	// Check if message already exist
	var messageID uuid.UUID

	row := tx.QueryRowContext(
		ctx,
		`SELECT id FROM message WHERE service_id = UUID_TO_BIN(?) AND language = ?`,
		serviceID,
		messages.Language.String(),
	)

	// Check if message already exists, if not, create a new one
	switch err = row.Scan(&messageID); {
	// Message already exists
	default:
		// noop

	// Message does not exist
	case errors.Is(err, sql.ErrNoRows):
		messageID = uuid.New()

		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO message (id, service_id, language) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?)`,
			messageID,
			serviceID,
			messages.Language.String(),
		); err != nil {
			return fmt.Errorf("repo: insert message: %w", err)
		}

	// Error scanning row
	case err != nil:
		return fmt.Errorf("repo: scan message: %w", err)
	}

	// Insert into message_message table,
	// on duplicate message_message.id and message_message.message_id,
	// update message's message, description and fuzzy values.
	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO message_message
	(message_id, id, message, description, fuzzy)
VALUES
	(UUID_TO_BIN(?), ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	message = VALUES(message),
	description = VALUES(description),
	fuzzy = VALUES(fuzzy)`,
	)
	if err != nil {
		return fmt.Errorf("repo: prepare stmt to insert message_message: %w", err)
	}
	defer stmt.Close()

	for _, m := range messages.Messages {
		_, err = stmt.ExecContext(
			ctx,
			messageID,
			m.ID,
			m.Message,
			m.Description,
			m.Fuzzy,
		)
		if err != nil {
			return &repo.DefaultError{Entity: "message_message", Err: err, Operation: "Insert"}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo: commit tx to save messages: %w", err)
	}

	return nil
}

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.Messages, error) {
	messages := &model.Messages{Language: language}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT mm.id, mm.message, mm.description, mm.fuzzy
FROM message_message mm
JOIN message m ON m.id = mm.message_id
WHERE m.service_id = UUID_TO_BIN(?) AND m.language = ?`,
		serviceID,
		language.String(),
	)
	if err != nil {
		return nil, &repo.DefaultError{Entity: "message_message", Err: err, Operation: "Select"}
	}

	defer rows.Close()

	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.Message, &m.Description, &m.Fuzzy); err != nil {
			return nil, &repo.DefaultError{Entity: "message_message", Err: err, Operation: "Scan"}
		}

		messages.Messages = append(messages.Messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, &repo.DefaultError{Entity: "message_message", Err: err, Operation: "Scan"}
	}

	return messages, nil
}
