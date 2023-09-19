package badgerdb

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
)

// Repo implements the repo interface.
type Repo struct {
	db *badger.DB
}

func (r *Repo) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close badger db: %w", err)
	}

	return nil
}

type option func(*Repo) error

// WithDefaultDB opens a new Badger database in file system
// with the path from global config e.g. ENV, flag or config file.
// If path is not provided defaults to in-memory storage.
func WithDefaultDB() option {
	return func(r *Repo) error {
		path := viper.GetString("db.badgerdb.path")
		badgerOpts := badger.DefaultOptions(path)

		// NOTE: The default value for in-memory storage of ValueThreshold is 1 MB.
		// Currently increasing the maximum allowed value size using WithValueThreshold() results in
		// panic 'Invalid ValueThreshold, must be less or equal to 1048576'.
		if path == "" {
			log.Println("INFO: badger db path not provided: defaulting to in-memory storage")

			badgerOpts = badgerOpts.WithInMemory(true)
		}

		var err error

		if r.db, err = newDB(badgerOpts); err != nil {
			return fmt.Errorf("WithDefaultDB: new badger db: %w", err)
		}

		return nil
	}
}

// NewRepo creates a new repo.
//
// NOTE: One of the options (WithDefaultDB or WithInMemoryDB) must be provided,
// otherwise it will return an error.
func NewRepo(opts ...option) (*Repo, error) {
	r := new(Repo)

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("apply option to repo: %w", err)
		}
	}

	return r, nil
}
