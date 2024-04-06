package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xo/dburl"
)

func connect(uri string) (*sql.DB, error) {
	parsed, err := dburl.Parse(uri)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", parsed.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
