package badgerdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

// SaveMessages handles both Create and Update.
func (r *Repo) SaveMessages(ctx context.Context, serviceID uuid.UUID, messages *model.Messages) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	err = r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("messages:%s:%s", serviceID, messages.Language))
		val, err := json.Marshal(messages)
		if err != nil {
			return fmt.Errorf("repo: marshal messages: %w", err)
		}
		return txn.Set(key, val)
	})
	if err != nil {
		return fmt.Errorf("repo: save messages: %w", err)
	}

	return nil
}

func (r *Repo) LoadMessages(ctx context.Context, serviceID uuid.UUID, language language.Tag) (*model.Messages, error) {
	messages := model.Messages{Language: language}

	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return &messages, nil
	}

	err = r.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("messages:%s:%s", serviceID, language))
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return fmt.Errorf("repo: get messages: %w", err)
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &messages)
		})
	})
	if err != nil {
		return nil, err
	}

	return &messages, nil
}
