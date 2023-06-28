package badgerdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo/common"

	"golang.org/x/text/language"
)

const messagesPrefix = "messages:"

// getMessagesKey converts a serviceID and language to a BadgerDB key with prefix.
func getMessagesKey(serviceID uuid.UUID, language language.Tag) []byte {
	return []byte(fmt.Sprintf("%s%s:%s", messagesPrefix, serviceID, language))
}

// SaveMessages handles both Create and Update.
func (r *Repo) SaveMessages(ctx context.Context, serviceID uuid.UUID, messages *model.Messages) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	err = r.db.Update(func(txn *badger.Txn) error {
		val, marshalErr := json.Marshal(messages)
		if marshalErr != nil {
			return fmt.Errorf("marshal messages: %w", err)
		}

		if setErr := txn.Set(getMessagesKey(serviceID, messages.Language), val); setErr != nil {
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
func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, opts common.LoadMessagesOpts,
) ([]model.Messages, error) {
	if _, err := r.LoadService(ctx, serviceID); errors.Is(err, common.ErrNotFound) {
		return nil, nil // Empty messages.messages for this service (Not an error)
	} else if err != nil {
		return nil, fmt.Errorf("repo: load service: %w", err)
	}

	// load all messages if language tags are not provided.
	if len(opts.FilterLanguages) == 0 {
		messages, err := r.loadMessages(serviceID)
		if err != nil {
			return nil, fmt.Errorf("load all messages for service '%s': %w", serviceID, err)
		}

		return messages, nil
	}

	// load messages based on provided language tags.
	messages, err := r.loadMessagesByLang(serviceID, opts.FilterLanguages)
	if err != nil {
		return nil, fmt.Errorf("load messages for language tags: %w", err)
	}

	return messages, nil
}

// loadMessagesByLang returns messages for service based on provided language tags.
func (r *Repo) loadMessagesByLang(serviceID uuid.UUID, languages []language.Tag,
) ([]model.Messages, error) {
	messages := make([]model.Messages, 0, len(languages))

	if err := r.db.View(func(txn *badger.Txn) error {
		for _, lang := range languages {
			var msgs model.Messages

			item, txErr := txn.Get(getMessagesKey(serviceID, lang))
			switch {
			default:
				if valErr := getValue(item, &msgs); valErr != nil {
					return fmt.Errorf("get messages for language tag '%s': %w", lang, valErr)
				}

				messages = append(messages, msgs)
			case errors.Is(txErr, badger.ErrKeyNotFound):
				return nil // Empty messages.messages for this language (Not an error)
			case txErr != nil:
				return fmt.Errorf("transaction: get messages for language tag '%s': %w", lang, txErr)
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return messages, nil
}

// loadMessages returns all messages for service.
func (r *Repo) loadMessages(serviceID uuid.UUID) ([]model.Messages, error) {
	keyPrefix := []byte(messagesPrefix + serviceID.String())

	var messages []model.Messages

	if err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(keyPrefix); it.ValidForPrefix(keyPrefix); it.Next() {
			msgs := model.Messages{}

			if err := getValue(it.Item(), &msgs); err != nil {
				return fmt.Errorf("transaction: get value: %w", err)
			}

			messages = append(messages, msgs)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return messages, nil
}
