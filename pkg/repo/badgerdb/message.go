package badgerdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
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

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.Messages, error) {
	messages := model.Messages{Language: language}

	_, err := r.LoadService(ctx, serviceID)

	switch {
	default:
	// noop
	case errors.Is(err, repo.ErrNotFound):
		return &messages, nil // Empty messages.messages for this service (Not an error)
	case err != nil:
		return nil, fmt.Errorf("repo: load service: %w", err)
	}

	err = r.db.View(func(txn *badger.Txn) error {
		item, getErr := txn.Get(getMessagesKey(serviceID, language))
		switch {
		default:
			// noop
		case errors.Is(getErr, badger.ErrKeyNotFound):
			return nil // Empty messages.messages for this language (Not an error)
		case getErr != nil:
			return fmt.Errorf("transaction: get messages: %w", getErr)
		}

		return getValues(item, &messages)
	})
	if err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return &messages, nil
}
