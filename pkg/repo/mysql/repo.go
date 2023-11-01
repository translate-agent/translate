package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type Repo struct {
	db *sql.DB
}

func (r *Repo) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close mysql db: %w", err)
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
		if r.db, err = NewDB(ctx, conf); err != nil {
			return fmt.Errorf("apply db conf to repo")
		}

		return nil
	}
}

func NewRepo(opts ...Option) (*Repo, error) {
	r := new(Repo)

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("apply option to repo: :%w", err)
		}
	}

	return r, nil
}

// eq returns empty squirrel.Eq if values is empty.
func eq[T any](column string, values []T) squirrel.Eq {
	if len(values) == 0 {
		return squirrel.Eq{}
	}

	return squirrel.Eq{column: values}
}
