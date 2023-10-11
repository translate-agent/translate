package factory

import (
	"context"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/repo/badgerdb"
	"go.expect.digital/translate/pkg/repo/mysql"
)

const (
	BadgerDB = "badgerdb"
	MySQL    = "mysql"
)

var SupportedDBs = []string{MySQL, BadgerDB}

// Usage returns a string describing the supported databases for CLI.
func Usage() string {
	return fmt.Sprintf("database to use. Supported options: %s", strings.Join(SupportedDBs, ", "))
}

// NewRepo creates a new repo based on the provided database string.
func NewRepo(ctx context.Context, db string) (repo.Repo, error) {
	var (
		repo repo.Repo
		err  error
	)

	switch v := strings.TrimSpace(strings.ToLower(db)); v {
	case SupportedDBs[0]: // mysql
		repo, err = mysql.NewRepo(mysql.WithDefaultDB(ctx))
	case SupportedDBs[1]: // badgerdb
		repo, err = badgerdb.NewRepo(badgerdb.WithDefaultDB())
	default:
		return nil, fmt.Errorf("unsupported database: '%s', list of supported db: %s", db, strings.Join(SupportedDBs, ", "))
	}

	if err != nil {
		return nil, fmt.Errorf("new '%s' repo: %w", db, err)
	}

	return repo, nil
}
