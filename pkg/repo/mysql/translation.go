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

func (r *Repo) SaveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error {
	if _, ok := r.db.(*sql.Tx); ok { // use existing tx
		return r.saveTranslation(ctx, serviceID, translation)
	}

	return r.Tx(ctx, func(ctx context.Context, rp repo.Repo) error { // create new tx
		return rp.SaveTranslation(ctx, serviceID, translation) //nolint:wrapcheck
	})
}

func (r *Repo) saveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	// Check if translation already exist
	var translationID uuid.UUID

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id FROM translation WHERE service_id = UUID_TO_BIN(?) AND language = ?`,
		serviceID,
		translation.Language.String(),
	)

	// Check if translation already exists, if not, create a new one
	switch err = row.Scan(&translationID); {
	// Translation already exists
	default:
		// noop

	// Translation does not exist
	case errors.Is(err, sql.ErrNoRows):
		translationID = uuid.New()

		if _, err = r.db.ExecContext(
			ctx,
			`INSERT INTO translation (id, service_id, language, original) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?)`,
			translationID,
			serviceID,
			translation.Language.String(),
			translation.Original,
		); err != nil {
			return fmt.Errorf("repo: insert message: %w", err)
		}

	// Error scanning row
	case err != nil:
		return fmt.Errorf("repo: scan message: %w", err)
	}

	// Insert into message table,
	// on duplicate message.id and message.translation_id,
	// update message's message, description and status values.
	stmt, err := r.db.PrepareContext(
		ctx,
		`INSERT INTO message
	(translation_id, id, message, description, positions, status)
VALUES
	(UUID_TO_BIN(?), ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	message = VALUES(message),
	description = VALUES(description),
	positions = VALUES(positions),
	status = VALUES(status)`,
	)
	if err != nil {
		return fmt.Errorf("repo: prepare stmt to insert message: %w", err)
	}
	defer stmt.Close()

	for _, m := range translation.Messages {
		_, err = stmt.ExecContext(
			ctx,
			translationID,
			m.ID,
			m.Message,
			m.Description,
			m.Positions,
			m.Status,
		)
		if err != nil {
			return fmt.Errorf("repo: insert message: %w", err)
		}
	}

	return nil
}

func (r *Repo) LoadTranslations(ctx context.Context, serviceID uuid.UUID, opts repo.LoadTranslationsOpts,
) (model.Translations, error) {
	rows, err := sq.
		Select("m.id, m.message, m.description, m.positions, m.status, t.language, t.original").
		From("message m").
		Join("translation t ON t.id = m.translation_id").
		Where("t.service_id = UUID_TO_BIN(?)", serviceID).
		Where(eq("t.language", langToStringSlice(opts.FilterLanguages))).
		RunWith(r.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("repo: query translations: %w", err)
	}

	defer rows.Close()

	translationsLookup := make(map[string]*model.Translation)

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

		// Lookup translation by language
		translation, ok := translationsLookup[lang]
		// If not found, create a new one
		if !ok {
			translation = &model.Translation{
				Language: language.MustParse(lang),
				Original: original,
			}
			// Add to lookup
			translationsLookup[lang] = translation
		}
		// Add scanned message to translation
		translation.Messages = append(translation.Messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: scan messages: %w", err)
	}

	allTranslations := make([]model.Translation, 0, len(translationsLookup))
	for _, translation := range translationsLookup {
		allTranslations = append(allTranslations, *translation)
	}

	return allTranslations, nil
}

// helpers

func langToStringSlice(languages []language.Tag) []string {
	lt := make([]string, 0, len(languages))
	for _, lang := range languages {
		lt = append(lt, lang.String())
	}

	return lt
}
