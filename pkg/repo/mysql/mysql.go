package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Conf struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func (c *Conf) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Password, c.Host, c.Port, c.Database)
}

func DefaultConf() (*Conf, error) {
	conf := &Conf{}

	if err := viper.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("viper: unmarshal to mysql conf: %w", err)
	}

	return conf, nil
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
