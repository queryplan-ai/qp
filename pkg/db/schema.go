package db

import (
	"net/url"
	"strings"

	"github.com/queryplan-ai/qp/pkg/db/types"
	"github.com/queryplan-ai/qp/pkg/mysql"
	"github.com/queryplan-ai/qp/pkg/pg"
	"github.com/xo/dburl"
)

func DatabaseNameFromURI(uri string) (string, error) {
	parsed, err := dburl.Parse(uri)
	if err != nil {
		return "", err
	}

	return strings.TrimLeft(parsed.Path, "/"), nil
}

func LoadSchema(db *types.DB) {
	switch dbEngine(db) {
	case "mysql":
		mysql.LoadSchema(db)
	case "postgres":
		pg.LoadSchema(db)
	}
}

func dbEngine(db *types.DB) string {
	uri, err := url.Parse(db.ConnectionURI)
	if err != nil {
		return ""
	}

	switch uri.Scheme {
	case "mysql":
		return "mysql"
	case "postgres", "postgresql":
		return "postgres"
	default:
		return ""
	}
}
