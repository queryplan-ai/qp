package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xo/dburl"
)

func VerifyMysqlConnection(uri string) (string, error) {
	parsed, err := dburl.Parse(uri)
	if err != nil {
		return "", err
	}

	db, err := sql.Open("mysql", parsed.DSN)
	if err != nil {
		return "", err
	}

	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return "", err
	}

	dbName, err := DatabaseNameFromURI(uri)
	if err != nil {
		return "", err
	}

	return dbName, nil
}
