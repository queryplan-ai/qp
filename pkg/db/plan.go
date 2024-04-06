package db

import (
	"github.com/queryplan-ai/qp/pkg/db/types"
	"github.com/queryplan-ai/qp/pkg/mysql"
)

func PlanQuery(db *types.DB, query string) (string, error) {
	switch dbEngine(db) {
	case "mysql":
		return mysql.PlanQuery(db, query)
	}

	return "", nil
}
