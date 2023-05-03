package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func (r *Repo) SaveMessages(ctx context.Context, serviceID uuid.UUID, messages *model.Messages) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo: begin tx to save messages: %w", err)
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

	err = row.Scan(&messageID)

	// If message does not exist, generate a new UUID for it
	switch {
	case errors.Is(err, sql.ErrNoRows):
		messageID = uuid.New()
	case err != nil:
		return fmt.Errorf("repo: scan message: %w", err)
	}

	// Insert into message, ignore if already exists
	_, err = tx.ExecContext(
		ctx,
		`INSERT IGNORE INTO message (id, service_id, language) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?)`,
		messageID,
		serviceID,
		messages.Language.String(),
	)
	if err != nil {
		return fmt.Errorf("repo: insert messages: %w", err)
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
		return fmt.Errorf("repo: prepare insert statement: %w", err)
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
			return fmt.Errorf("repo: insert message_message: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo: commit tx to save messages: %w", err)
	}

	return nil
}

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.Messages, error) {
	var (
		messages  = &model.Messages{Language: language}
		messageID uuid.UUID
	)

	// Check if message with service_id and language exists
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id FROM message WHERE service_id = UUID_TO_BIN(?) AND language = ?`,
		serviceID,
		language.String(),
	)

	// If message does not exist, return empty messages
	switch err := row.Scan(&messageID); {
	case errors.Is(err, sql.ErrNoRows):
		return messages, nil
	case err != nil:
		return nil, fmt.Errorf("repo: scan message: %w", err)
	}

	// Load messages
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, message, description, fuzzy
FROM message_message
WHERE message_id = UUID_TO_BIN(?)`,
		messageID,
	)
	if err != nil {
		return nil, fmt.Errorf("repo: query messages: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.Message, &m.Description, &m.Fuzzy); err != nil {
			return nil, fmt.Errorf("repo: scan message: %w", err)
		}

		messages.Messages = append(messages.Messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: scan messages: %w", err)
	}

	return messages, nil
}
