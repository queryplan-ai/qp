package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func connect(uri string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), uri)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return conn, nil
}
