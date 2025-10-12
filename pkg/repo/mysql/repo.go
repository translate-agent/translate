package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"go.expect.digital/translate/pkg/repo"
)

// DB interface defines method signatures found both in sql.DB and sql.Tx.
type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type Repo struct {
	db DB
}

func NewRepo(opts ...Option) (*Repo, error) {
	r := new(Repo)

	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, fmt.Errorf("apply option to repo: :%w", err)
		}
	}

	return r, nil
}

func (r *Repo) Close() error {
	if db, ok := r.db.(*sql.DB); ok {
		err := db.Close()
		if err != nil {
			return fmt.Errorf("close mysql db: %w", err)
		}
	}

	return nil
}

// Option function used for setting optional Repo properties.
type Option func(*Repo) error

func WithDB(db *sql.DB) Option {
	return func(r *Repo) error {
		r.db = db

		return nil
	}
}

// WithDefaultDB reads configuration data from Viper and uses it to create a new DB.
func WithDefaultDB(ctx context.Context) Option {
	return func(r *Repo) (err error) {
		conf := DefaultConf()

		r.db, err = NewDB(ctx, conf)
		if err != nil {
			return fmt.Errorf("connect to DB from default conf: %w", err)
		}

		return nil
	}
}

func WithConf(ctx context.Context, conf *Conf) Option {
	return func(r *Repo) (err error) {
		r.db, err = NewDB(ctx, conf)
		if err != nil {
			return errors.New("apply db conf to repo")
		}

		return nil
	}
}

// eq returns empty squirrel.Eq if values is empty.
func eq[T any](column string, values []T) squirrel.Eq {
	if len(values) == 0 {
		return squirrel.Eq{}
	}

	return squirrel.Eq{column: values}
}

// Tx executes a transaction on the repository.
// Returns an error if there was an error starting the transaction,
// executing the callback function, or committing the transaction.
func (r *Repo) Tx(ctx context.Context, fn func(context.Context, repo.Repo) error) (err error) {
	var tx *sql.Tx

	switch db := r.db.(type) {
	case *sql.DB: // start transaction
		tx, err = db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("repo: begin tx: %w", err)
		}
	case *sql.Tx:
		return errors.New("repo: tx already exists")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() //nolint:errcheck

			err = fmt.Errorf("repo: tx panicked: %v", r)
		}
	}()

	err = fn(ctx, &Repo{db: tx})
	if err != nil {
		tx.Rollback() //nolint:errcheck

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
	switch db := r.db.(type) {
	case *sql.Tx: // use existing tx
		return fn(ctx, r)
	case *sql.DB:
		var tx *sql.Tx

		tx, err = db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("repo: begin tx: %w", err)
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback() //nolint:errcheck

				err = fmt.Errorf("repo: tx panicked: %v", r)
			}
		}()

		err = fn(ctx, &Repo{db: tx})
		if err != nil {
			tx.Rollback() //nolint:errcheck

			return fmt.Errorf("repo: execute tx: %w", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("repo: commit tx: %w", err)
		}
	}

	return nil
}
