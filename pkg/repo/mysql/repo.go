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

// WithDefaultDB reads configuration data from Viper and uses it to create a new DB.
func WithDefaultDB(ctx context.Context) Option {
	return optionFunc(func(r *Repo) (err error) {
		conf := DefaultConf()

		r.db, err = NewDB(ctx, conf)
		if err != nil {
			return fmt.Errorf("create new db from conf: %w", err)
		}

		return nil
	})
}

func WithConf(ctx context.Context, conf *Conf) Option {
	return optionFunc(func(r *Repo) (err error) {
		if r.db, err = NewDB(ctx, conf); err != nil {
			return fmt.Errorf("apply db conf to repo")
		}
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
