package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/queryplan-ai/qp/pkg/db/types"
)

func VerifyPGConnection(uri string) (string, error) {
	conn, err := pgx.Connect(context.Background(), uri)
	if err != nil {
		return "", err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return "", err
	}

	query := `select version()`
	row := conn.QueryRow(context.Background(), query)
	var reportedVersion string
	if err := row.Scan(&reportedVersion); err != nil {
		return "", err
	}

	dbName, err := DatabaseNameFromURI(uri)
	if err != nil {
		return "", err
	}

	return dbName, nil
}

func loadPostgresSchema(db *types.DB) {

}
