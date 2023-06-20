package repo

import (
	"context"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/repo/badgerdb"
	"go.expect.digital/translate/pkg/repo/mysql"
)

type RepoDB int

const (
	MySQL RepoDB = iota
	BadgerDB
)

var RepoDBNames = map[RepoDB]string{
	MySQL:    "mysql",
	BadgerDB: "badgerdb",
}

func (r RepoDB) String() string {
	return RepoDBNames[r]
}

// SupportedDBs returns a list of supported databases.
func SupportedDBs() []string {
	dbs := make([]string, 0, len(RepoDBNames))
	for _, v := range RepoDBNames {
		dbs = append(dbs, v)
	}

	return dbs
}

// Usage returns a string describing the supported databases for CLI.
func Usage() string {
	return fmt.Sprintf("database to use. Supported options: %s", strings.Join(SupportedDBs(), ", "))
}

// parseRepoDB parses the string representation of a database and returns the corresponding RepoDB.
func parseRepoDB(db string) (RepoDB, error) {
	switch v := strings.TrimSpace(strings.ToLower(db)); v {
	case "mysql":
		return MySQL, nil
	case "badgerdb":
		return BadgerDB, nil
	default:
		return -1, fmt.Errorf("unsupported database: '%s', list of supported db: %s", db, strings.Join(SupportedDBs(), ", "))
	}
}

// NewRepo creates a new repo based on the provided database string.
func NewRepo(ctx context.Context, db string) (Repo, error) {
	dbType, err := parseRepoDB(db)
	if err != nil {
		return nil, fmt.Errorf("parse repo db: %w", err)
	}

	var repo Repo

	switch dbType {
	case MySQL:
		repo, err = mysql.NewRepo(mysql.WithDefaultDB(ctx))
	case BadgerDB:
		repo, err = badgerdb.NewRepo(badgerdb.WithDefaultDB())
	}

	if err != nil {
		return nil, fmt.Errorf("new '%s' repo: %w", dbType, err)
	}

	return repo, nil
}
