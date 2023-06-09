package badgerdb

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

func NewDB(options badger.Options) (*badger.DB, error) {
	db, err := badger.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open badger db: %w", err)
	}

	// Check if works

	// Start a read-only transaction.
	tx := db.NewTransaction(false)
	defer tx.Discard()

	// Get the value for the key "answer".
	_, err = tx.Get([]byte("Hello"))

	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, fmt.Errorf("ping badger db: %w", err)
	}

	return db, nil
}
