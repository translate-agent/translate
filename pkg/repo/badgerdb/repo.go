package badgerdb

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

// Repo implements the repo interface.
type Repo struct {
	db *badger.DB
}

type option func(*Repo) error

// WithDefaultDB opens a new Badger database in file system
// with the path from global config e.g. ENV, flag or config file.
func WithDefaultDB() option {
	return func(r *Repo) error {
		path := viper.GetString("db.badgerdb.path")
		if path == "" {
			// TODO: tidy
			log.Println("badger db path is not set using in memory")

			var err error
			if r.db, err = newDB(badger.DefaultOptions("").WithInMemory(true)); err != nil {
				return fmt.Errorf("WithInMemoryDB: open badger db: %w", err)
			}

			return nil
		}

		var err error

		r.db, err = newDB(badger.DefaultOptions(path))
		if err != nil {
			return fmt.Errorf("WithDefaultDB: open badger db: %w", err)
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
