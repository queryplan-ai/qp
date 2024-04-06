package shell

import (
	"fmt"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/queryplan-ai/qp/pkg/db"
	"github.com/queryplan-ai/qp/pkg/shell/types"
)

func handleQuery(sh *types.Shell, query string) *types.ShellCommandResult {
	result := types.ShellCommandResult{
		IsFatal:   false,
		IsSuccess: false,
	}

	if !isQuery(query) {
		result.Message = "not a valid query"
		return &result
	}

	message, err := db.PlanQuery(sh.DB, query)
	if err != nil {
		result.Message = fmt.Sprintf("Error planning query: %s", err)
		return &result
	}

	result.IsSuccess = true
	result.Message = message

	return &result
}

func isQuery(query string) bool {
	defer func() {
		recover()
	}()

	// a database query is a string that starts with "select", "insert", "update", "delete"
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		fmt.Printf("Error parsing query: %s", err)
		return false
	}

	switch stmt.(type) {
	case *sqlparser.Select, *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
		return true
	default:
		return false
	}
}
