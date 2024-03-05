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
)

const servicePrefix = "service:"

// getServiceKey converts a service ID to a BadgerDB key with prefix.
func getServiceKey(id uuid.UUID) []byte {
	return []byte(fmt.Sprintf("%s%s", servicePrefix, id))
}

func (r *Repo) SaveService(ctx context.Context, service *model.Service) error {
	if service.ID == uuid.Nil {
		service.ID = uuid.New()
	}

	err := r.db.Update(func(txn *badger.Txn) error {
		val, err := json.Marshal(service)
		if err != nil {
			return fmt.Errorf("marshal service: %w", err)
		}

		if err := txn.Set(getServiceKey(service.ID), val); err != nil {
			return fmt.Errorf("transaction: set service: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("repo: db update: %w", err)
	}

	return nil
}

func (r *Repo) LoadService(ctx context.Context, serviceID uuid.UUID) (*model.Service, error) {
	var service model.Service

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getServiceKey(serviceID))

		switch {
		default:
			return getValue(item, &service)
		case errors.Is(err, badger.ErrKeyNotFound):
			return repo.ErrNotFound
		case err != nil:
			return fmt.Errorf("transaction: get service: %w", err)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return &service, nil
}

func (r *Repo) LoadServices(ctx context.Context) ([]model.Service, error) {
	var services []model.Service

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(servicePrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			var service model.Service

			if err := getValue(item, &service); err != nil {
				return err
			}

			services = append(services, service)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return services, nil
}

func (r *Repo) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	key := getServiceKey(serviceID)

	err := r.db.Update(func(txn *badger.Txn) error {
		// BadgerDB does not return an error if the key does not exist on Delete.
		// So we have to check if the key exists first.
		_, err := txn.Get(key)

		switch {
		default:
			if deleteErr := txn.Delete(key); deleteErr != nil {
				return fmt.Errorf("transaction: delete service: %w", deleteErr)
			}

			return nil
		case errors.Is(err, badger.ErrKeyNotFound):
			return repo.ErrNotFound
		case err != nil:
			return fmt.Errorf("transaction: get service for deletion: %w", err)
		}
	})
	if err != nil {
		return fmt.Errorf("repo: db update: %w", err)
	}

	return nil
}
