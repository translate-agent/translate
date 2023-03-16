package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"go.nhat.io/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.14.0"
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
	return new(Conf), nil
}

func NewDB(ctx context.Context, conf *Conf) (*sql.DB, error) {
	// https://github.com/nhatthm/otelsql
	driverName, err := otelsql.Register("mysql",
		otelsql.AllowRoot(), // For integration tests.
		otelsql.TraceQueryWithArgs(),
		otelsql.WithDatabaseName(conf.Database),
		otelsql.WithSystem(semconv.DBSystemMySQL),
	)
	if err != nil {
		return nil, fmt.Errorf("register driver: %w", err)
	}

	db, err := sql.Open(driverName, conf.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("connect to MySQL: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}

	return db, nil
}
