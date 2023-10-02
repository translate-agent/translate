package badgerdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
)

const messagesPrefix = "translation:"

// messagesKey converts a serviceID and language to a BadgerDB key with prefix.
func messagesKey(serviceID uuid.UUID, language language.Tag) []byte {
	return []byte(fmt.Sprintf("%s%s:%s", messagesPrefix, serviceID, language))
}

// SaveMessages handles both Create and Update.
func (r *Repo) SaveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	err = r.db.Update(func(txn *badger.Txn) error {
		val, marshalErr := json.Marshal(translation)
		if marshalErr != nil {
			return fmt.Errorf("marshal messages: %w", err)
		}

		if setErr := txn.Set(messagesKey(serviceID, translation.Language), val); setErr != nil {
			return fmt.Errorf("transaction: set messages: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("repo: db update: %w", err)
	}

	return nil
}

// LoadMessages retrieves messages from db based on serviceID and LoadMessageOpts.
func (r *Repo) LoadTranslation(ctx context.Context, serviceID uuid.UUID, opts repo.LoadTranslationOpts,
) (model.TranslationSlice, error) {
	if _, err := r.LoadService(ctx, serviceID); errors.Is(err, repo.ErrNotFound) {
		return nil, nil // Empty translation.messages for this service (Not an error)
	} else if err != nil {
		return nil, fmt.Errorf("repo: load service: %w", err)
	}

	// load all messages if languages are not provided.
	if len(opts.FilterLanguages) == 0 {
		messages, err := r.loadMessages(serviceID)
		if err != nil {
			return nil, fmt.Errorf("load messages by service '%s': %w", serviceID, err)
		}

		return messages, nil
	}

	// load messages based on provided languages.
	messages, err := r.loadMessagesByLang(serviceID, opts.FilterLanguages)
	if err != nil {
		return nil, fmt.Errorf("load messages by languages: %w", err)
	}

	return messages, nil
}

// loadMessagesByLang returns messages for service based on provided languages.
func (r *Repo) loadMessagesByLang(serviceID uuid.UUID, languages []language.Tag,
) (model.TranslationSlice, error) {
	translations := make([]model.Translation, 0, len(languages))

	if err := r.db.View(func(txn *badger.Txn) error {
		for _, lang := range languages {
			var translation model.Translation

			item, txErr := txn.Get(messagesKey(serviceID, lang))
			switch {
			default:
				if valErr := getValue(item, &translation); valErr != nil {
					return fmt.Errorf("get translations by language '%s': %w", lang, valErr)
				}

				translations = append(translations, translation)
			case errors.Is(txErr, badger.ErrKeyNotFound):
				return nil // Empty translations.translations for this language (Not an error)
			case txErr != nil:
				return fmt.Errorf("transaction: get translations by language '%s': %w", lang, txErr)
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return translations, nil
}

// loadMessages returns all messages for service.
func (r *Repo) loadMessages(serviceID uuid.UUID) (model.TranslationSlice, error) {
	keyPrefix := []byte(messagesPrefix + serviceID.String())

	var translations []model.Translation

	if err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(keyPrefix); it.ValidForPrefix(keyPrefix); it.Next() {
			translation := model.Translation{}

			if err := getValue(it.Item(), &translation); err != nil {
				return fmt.Errorf("transaction: get value: %w", err)
			}

			translations = append(translations, translation)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return translations, nil
}
