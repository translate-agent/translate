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

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, opts common.LoadMessagesOpts,
) ([]model.Messages, error) {
	var err error

	messages := make([]model.Messages, 0, len(opts.FilterLanguages))

	// load messages based on provided language tags.
	if len(opts.FilterLanguages) > 0 {
		for _, langTag := range opts.FilterLanguages {
			msgs, er := r.LoadIndividualMessages(ctx, serviceID, langTag)
			if er != nil {
				return nil, fmt.Errorf("load messages for service '%s' language '%s': %w", serviceID, langTag, er)
			}

			if msgs != nil && len(msgs.Messages) != 0 {
				messages = append(messages, *msgs)
			}
		}

		return messages, nil
	}

	messages, err = r.LoadAllMessages(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("load all messages for service '%s': %w", serviceID, err)
	}

	return messages, nil
}

// LoadIndividualMessages returns messages based on serviceID and language.Tag.
func (r *Repo) LoadIndividualMessages(ctx context.Context, serviceID uuid.UUID, language language.Tag,
) (*model.Messages, error) {
	var messages model.Messages

	_, err := r.LoadService(ctx, serviceID)

	switch {
	default:
	// noop
	case errors.Is(err, common.ErrNotFound):
		return &messages, nil // Empty messages.messages for this service (Not an error)
	case err != nil:
		return nil, fmt.Errorf("repo: load service: %w", err)
	}

	err = r.db.View(func(txn *badger.Txn) error {
		item, getErr := txn.Get(getMessagesKey(serviceID, language))
		switch {
		default:
			return getValue(item, &messages)
		case errors.Is(getErr, badger.ErrKeyNotFound):
			return nil // Empty messages.messages for this language (Not an error)
		case getErr != nil:
			return fmt.Errorf("transaction: get messages: %w", getErr)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return &messages, nil
}

// LoadAllMessages returns all messages for service.
func (r *Repo) LoadAllMessages(ctx context.Context, serviceID uuid.UUID) ([]model.Messages, error) {
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
