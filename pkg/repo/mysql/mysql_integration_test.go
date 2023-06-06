//go:build integration

package mysql

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

var repository *Repo

func TestMain(m *testing.M) {
	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	var err error

	repository, err = NewRepo(WithDefaultDB(ctx))
	if err != nil {
		log.Panicf("create new repo: %v", err)
	}

	code := m.Run()

	repository.db.Close()

	os.Exit(code)
}
