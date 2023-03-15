package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Conf struct {
	Addr     string
	User     string
	Password string
	Database string
}

func DefaultConf() (*Conf, error) {
	return new(Conf), nil
}

func NewDB(ctx context.Context, conf *Conf) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", conf.User, conf.Password, conf.Addr, conf.Database))
	if err != nil {
		return nil, fmt.Errorf("connect to MySQL: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}

	return db, nil
}
