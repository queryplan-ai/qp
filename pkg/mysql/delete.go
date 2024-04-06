package mysql

import (
	"fmt"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	issuetypes "github.com/queryplan-ai/qp/pkg/issue/types"
)

func scanDeleteStatementForIssues(query string, mysqlTables []MysqlTable) ([]issuetypes.QueryIssue, error) {
	_, err := parseDeleteStatement(query)
	if err != nil {
		return nil, fmt.Errorf("parse delete statement: %w", err)
	}

	return nil, nil
}
func parseDeleteStatement(query string) (*DeleteStatement, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse delete statement: %w", err)
	}

	deleteStmt, ok := stmt.(*sqlparser.Delete)
	if !ok {
		return nil, fmt.Errorf("parse delete statement: not a delete statement")
	}

	result := DeleteStatement{
		Tables: []string{},
	}

	// Extract table names
	result.Tables = extractDeleteTableNames(deleteStmt)

	return &result, nil
}

type DeleteStatement struct {
	Tables []string // List of tables being deleted from
}

func extractDeleteTableNames(deleteStmt *sqlparser.Delete) []string {
	var tables []string
	for _, tableExpr := range deleteStmt.TableExprs {
		switch expr := tableExpr.(type) {
		case *sqlparser.AliasedTableExpr:
			tableName := sqlparser.String(expr.Expr)
			tables = append(tables, tableName)
		}
	}
	return tables
}
