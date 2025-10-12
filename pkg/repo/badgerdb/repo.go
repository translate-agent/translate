package badgerdb

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/repo"
)

// Repo implements the repo interface.
type Repo struct {
	db *badger.DB
	tx *badger.Txn
}

// NewRepo creates a new repo.
//
// NOTE: One of the options (WithDefaultDB or WithInMemoryDB) must be provided,
// otherwise it will return an error.
func NewRepo(opts ...Option) (*Repo, error) {
	r := new(Repo)

	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, fmt.Errorf("apply option to repo: %w", err)
		}
	}

	return r, nil
}

func (r *Repo) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close badger db: %w", err)
	}

	return nil
}

type Option func(*Repo) error

// WithDefaultDB opens a new Badger database in file system
// with the path from global config e.g. ENV, flag or config file.
// If path is not provided defaults to in-memory storage.
func WithDefaultDB() Option {
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

		r.db, err = newDB(badgerOpts)
		if err != nil {
			return fmt.Errorf("WithDefaultDB: new badger db: %w", err)
		}

		return nil
	}
}

func (r *Repo) Tx(ctx context.Context, fn func(context.Context, repo.Repo) error) (err error) {
	if r.tx != nil {
		return errors.New("repo: tx already exists")
	}

	tx := r.db.NewTransaction(true)

	defer func() {
		if r := recover(); r != nil {
			tx.Discard()

			err = fmt.Errorf("repo: tx panicked: %v", r)
		}
	}()

	err = fn(ctx, &Repo{db: r.db, tx: tx})
	if err != nil {
		tx.Discard()

		return fmt.Errorf("repo: execute tx: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("repo: commit tx: %w", err)
	}

	return nil
}

// ensureTx checks for existing db transaction - if present uses existing, otherwise starts a new tx.
func (r *Repo) ensureTx(ctx context.Context, fn func(context.Context, *Repo) error) (err error) {
	if r.tx != nil { // use existing tx
		return fn(ctx, r)
	}

	tx := r.db.NewTransaction(true)

	defer func() {
		if r := recover(); r != nil {
			tx.Discard()

			err = fmt.Errorf("repo: tx panicked: %v", r)
		}
	}()

	err = fn(ctx, &Repo{db: r.db, tx: tx})
	if err != nil {
		tx.Discard()

		return fmt.Errorf("repo: execute tx: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("repo: commit tx: %w", err)
	}

	return nil
}
