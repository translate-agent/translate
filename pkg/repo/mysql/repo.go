package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

// sq is SQL builder for MySQL.
var sq squirrel.StatementBuilderType

func init() {
	sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
}

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
			return fmt.Errorf("connect to DB from default conf: %w", err)
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

// helpers - SQL builder.

type eb squirrel.Eq

// in adds where clause only if string values are not empty.
func (e eb) in(column string, values []string) eb {
	if len(values) > 0 {
		e[column] = values
	}

	return e
}

// eq returns squirrel.Eq.
func (e eb) eq() squirrel.Eq {
	return squirrel.Eq(e)
}
