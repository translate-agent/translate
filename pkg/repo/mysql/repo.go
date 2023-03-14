package mysql

import (
	"context"
	"database/sql"
	"fmt"
)

type Repo struct {
	db *sql.DB
}

// Option interface used for setting optional Repo properties.
type Option interface {
	apply(*Repo) error
}

type optionFunc func(*Repo) error

func (o optionFunc) apply(c *Repo) error { return o(c) }

func WithDB(db *sql.DB) Option {
	return optionFunc(func(r *Repo) error {
		r.db = db

		return nil
	})
}

func WithDefaultDB(ctx context.Context) Option {
	return optionFunc(func(r *Repo) error {
		conf, err := DefaultConf()
		if err != nil {
			return fmt.Errorf("apply default db conf to repo: %w", err)
		}

		db, err := NewDB(ctx, conf)
		if err != nil {
			return fmt.Errorf("apply default db to repo: %w", err)
		}

		r.db = db

		return nil
	})
}

func NewRepo(opts ...Option) (*Repo, error) {
	r := new(Repo)

	for _, opt := range opts {
		if err := opt.apply(r); err != nil {
			return nil, fmt.Errorf("apply option to repo: :%w", err)
		}
	}

	return r, nil
}
