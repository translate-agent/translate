//go:build integration

package factory

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/repo/badgerdb"
	"go.expect.digital/translate/pkg/repo/mysql"
	"go.expect.digital/translate/pkg/testutil"
)

// repos is a map of all possible repo with different backends. E.g. MySQL, BadgerDB, etc.
var repos map[string]repo.Repo

// initMysql creates a new MySQL repo and adds it to the repos map.
func initMysql(ctx context.Context) error {
	repo, err := mysql.NewRepo(mysql.WithDefaultDB(ctx))
	if err != nil {
		return fmt.Errorf("create new mysql repo: %w", err)
	}

	repos["MySQL"] = repo

	return nil
}

// initBadgerDB creates a new BadgerDB repo and adds it to the repos map.
func initBadgerDB() error {
	repo, err := badgerdb.NewRepo(badgerdb.WithDefaultDB())
	if err != nil {
		return fmt.Errorf("create new badgerdb repo: %w", err)
	}

	repos["BadgerDB"] = repo

	return nil
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	// Initialize repos
	repos = make(map[string]repo.Repo, len(SupportedDBs))

	// MySQL
	if err := initMysql(ctx); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// BadgerDB
	if err := initBadgerDB(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Close all connections
	for _, repo := range repos {
		defer repo.Close()
	}

	return m.Run()
}

// allRepos runs a test for each repo that is defined in the repos map.
func allRepos(t *testing.T, f func(t *testing.T, repo repo.Repo, subtest testutil.SubtestFn)) {
	for name, repo := range repos {
		name, repo := name, repo
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, subTest := testutil.Trace(t)

			f(t, repo, subTest)
		})
	}
}
