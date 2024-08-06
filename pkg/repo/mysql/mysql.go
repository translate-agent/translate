package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/spf13/viper"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Conf struct {
	Host     string
	User     string
	Password string
	Database string
	Port     int
}

func (c *Conf) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.User, c.Password, c.Host, c.Port, c.Database)
}

func DefaultConf() *Conf {
	return &Conf{
		Host:     viper.GetString("db.mysql.host"),
		Port:     viper.GetInt("db.mysql.port"),
		User:     viper.GetString("db.mysql.user"),
		Password: viper.GetString("db.mysql.password"),
		Database: viper.GetString("db.mysql.database"),
	}
}

func NewDB(ctx context.Context, conf *Conf) (*sql.DB, error) {
	// https://github.com/XSAM/otelsql
	db, err := otelsql.Open(
		"mysql",
		conf.ConnectionString(),
		otelsql.WithAttributes(
			semconv.DBSystemMySQL,
			semconv.DBNameKey.String(conf.Database),
		),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			Ping:                 false,
			RowsNext:             false,
			DisableErrSkip:       true,
			DisableQuery:         false,
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitConnQuery:        true,
			OmitRows:             true,
			OmitConnectorConnect: true,
		}))
	if err != nil {
		return nil, fmt.Errorf("connect to MySQL: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}

	return db, nil
}
