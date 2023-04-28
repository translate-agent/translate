package mysql

import (
	"context"
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
		return fmt.Errorf("repo: begin tx: %w", err)
	}

	defer tx.Rollback() //nolint:errcheck

	// Insert into messages table
	_, err = tx.ExecContext(
		ctx,
		`INSERT IGNORE INTO message (service_id, language) VALUES (UUID_TO_BIN(?), ?)`,
		serviceID.String(),
		messages.Language.String(),
	)
	if err != nil {
		return fmt.Errorf("repo: insert messages: %w", err)
	}

	// Insert into message_message table
	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO message_message
	(id, message_service_id, message_language, message_id, message, description, fuzzy)
VALUES
	(UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?, ?, ?, ?)
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
			uuid.New().String(),
			serviceID.String(),
			messages.Language.String(),
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
		return fmt.Errorf("repo: commit tx: %w", err)
	}

	return nil
}

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.Messages, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT message_id, message, description, fuzzy
FROM message_message
WHERE message_service_id = UUID_TO_BIN(?)
AND message_language = ?`,
		serviceID.String(),
		language.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("repo: query messages: %w", err)
	}

	defer rows.Close()

	var messages []model.Message

	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.Message, &m.Description, &m.Fuzzy); err != nil {
			return nil, fmt.Errorf("repo: scan message: %w", err)
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: scan messages: %w", err)
	}

	return &model.Messages{Language: language, Messages: messages}, nil
}
