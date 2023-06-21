package badgerdb

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"go.expect.digital/translate/pkg/model"
)

// newDB opens a new Badger database with the given options.
func newDB(options badger.Options) (*badger.DB, error) {
	db, err := badger.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open badger db: %w", err)
	}

	// Check if the database was opened successfully by doing a simple Get.

	// Start a read-only transaction.
	tx := db.NewTransaction(false)
	defer tx.Discard()

	_, err = tx.Get([]byte("Hello"))

	// If the error is something other than "Key not found", something is wrong.
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, fmt.Errorf("ping badger db: %w", err)
	}

	return db, nil
}

// getValue unmarshals the value of a BadgerDB item into v (either *model.Service or *model.Messages).
func getValue[T *model.Service | *model.Messages](item *badger.Item, v T) error {
	return item.Value(func(val []byte) error { //nolint:wrapcheck
		if err := json.Unmarshal(val, &v); err != nil {
			return fmt.Errorf("unmarshal %T: %w", v, err)
		}
		return nil
	})
}
