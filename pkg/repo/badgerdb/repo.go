package badgerdb

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

type Repo struct {
	db *badger.DB
}

type option func(*Repo) error

func WithDefaultDB() option {
	return func(r *Repo) error {
		path := viper.GetString("db.badgerdb.path")
		if path == "" {
			return fmt.Errorf("WithDefaultDB: viper: empty db.badgerdb.path")
		}

		var err error
		r.db, err = NewDB(badger.DefaultOptions(path))

		if err != nil {
			return fmt.Errorf("WithDefaultDB: open badger db: %w", err)
		}

		return nil
	}
}

func WithInMemoryDB() option {
	return func(r *Repo) error {
		var err error

		r.db, err = NewDB(badger.DefaultOptions("").WithInMemory(true))
		if err != nil {
			return fmt.Errorf("WithInMemoryDB: open badger db: %w", err)
		}

		return nil
	}
}

func NewRepo(opts ...option) (*Repo, error) {
	r := new(Repo)

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("apply option to repo: %w", err)
		}
	}

	return r, nil
}
