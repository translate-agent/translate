package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo/common"
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
	// update message's message, description and status values.
	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO message_message
	(message_id, id, message, description, status)
VALUES
	(UUID_TO_BIN(?), ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	message = VALUES(message),
	description = VALUES(description),
	status = VALUES(status)`,
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
			m.Status,
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

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, opts common.LoadMessagesOpts,
) ([]model.Messages, error) {
	rows, err := sq.
		Select("mm.id, mm.message, mm.description, mm.status, m.language").
		From("message_message mm").
		Join("message m ON m.id = mm.message_id").
		Where("m.service_id = UUID_TO_BIN(?)", serviceID).
		Where(eq("m.language", langToStringSlice(opts.FilterLanguages))).
		RunWith(r.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("repo: query messages: %w", err)
	}

	defer rows.Close()

	messagesLookup := make(map[string][]model.Message)

	for rows.Next() {
		var (
			msg  model.Message
			lang string
		)

		if err := rows.Scan(&msg.ID, &msg.Message, &msg.Description, &msg.Status, &lang); err != nil {
			return nil, fmt.Errorf("repo: scan message: %w", err)
		}

		messagesLookup[lang] = append(messagesLookup[lang], msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: scan messages: %w", err)
	}

	messages := make([]model.Messages, 0, len(messagesLookup))

	for lang, msgs := range messagesLookup {
		messages = append(messages, model.Messages{
			Language: language.MustParse(lang),
			Messages: msgs,
		})
	}

	return messages, nil
}

// helpers

func langToStringSlice(languages []language.Tag) []string {
	lt := make([]string, 0, len(languages))
	for _, lang := range languages {
		lt = append(lt, lang.String())
	}

	return lt
}
