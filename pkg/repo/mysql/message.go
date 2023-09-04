package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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
			`INSERT INTO message (id, service_id, language, original) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?)`,
			messageID,
			serviceID,
			messages.Language.String(),
			messages.Original,
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
	(message_id, id, message, description, positions, status)
VALUES
	(UUID_TO_BIN(?), ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	message = VALUES(message),
	description = VALUES(description),
	positions = VALUES(positions),
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
			m.Positions,
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

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, opts repo.LoadMessagesOpts,
) (model.MessagesSlice, error) {
	rows, err := sq.
		Select("mm.id, mm.message, mm.description, mm.positions, mm.status, m.language, m.original").
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

	messagesLookup := make(map[string]*model.Messages)

	for rows.Next() {
		var (
			msg      model.Message
			lang     string
			original bool
		)

		if err := rows.Scan(
			&msg.ID, &msg.Message, &msg.Description, &msg.Positions, &msg.Status, &lang, &original); err != nil {
			return nil, fmt.Errorf("repo: scan message: %w", err)
		}

		// Lookup message by language
		messages, ok := messagesLookup[lang]
		// If not found, create a new one
		if !ok {
			messages = &model.Messages{
				Language: language.MustParse(lang),
				Original: original,
			}
			// Add to lookup
			messagesLookup[lang] = messages
		}
		// Add scanned message to messages
		messages.Messages = append(messages.Messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: scan messages: %w", err)
	}

	allMessages := make([]model.Messages, 0, len(messagesLookup))
	for _, msgs := range messagesLookup {
		allMessages = append(allMessages, *msgs)
	}

	return allMessages, nil
}

// helpers

func langToStringSlice(languages []language.Tag) []string {
	lt := make([]string, 0, len(languages))
	for _, lang := range languages {
		lt = append(lt, lang.String())
	}

	return lt
}
