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
)

func (r *Repo) SaveService(ctx context.Context, service *model.Service) error {
	if service.ID == uuid.Nil {
		service.ID = uuid.New()
	}

	err := r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("service:%s", service.ID))
		val, err := json.Marshal(service)
		if err != nil {
			return fmt.Errorf("repo: marshal service: %w", err)
		}
		return txn.Set(key, val)
	})
	if err != nil {
		return fmt.Errorf("repo: insert service: %w", err)
	}

	return nil
}

func (r *Repo) LoadService(ctx context.Context, serviceID uuid.UUID) (*model.Service, error) {
	var service model.Service

	err := r.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("service:%s", serviceID))
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return repo.ErrNotFound
			}
			return fmt.Errorf("repo: get service: %w", err)
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &service)
		})
	})
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *Repo) LoadServices(ctx context.Context) ([]model.Service, error) {
	var services []model.Service

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("service:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var service model.Service
				if err := json.Unmarshal(val, &service); err != nil {
					return fmt.Errorf("repo: unmarshal service: %w", err)
				}
				services = append(services, service)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (r *Repo) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	key := []byte(fmt.Sprintf("service:%s", serviceID))

	tx := r.db.NewTransaction(true)
	defer tx.Commit()

	err := r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		switch {
		default:
			return txn.Delete(key)
		case errors.Is(err, badger.ErrKeyNotFound):
			return repo.ErrNotFound
		case err != nil:
			return fmt.Errorf("repo: get service: %w", err)
		}
	})

	fmt.Printf("err: %v\n", err)

	if errors.Is(err, badger.ErrKeyNotFound) {
		return repo.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("repo: delete service: %w", err)
	}

	return nil
}
