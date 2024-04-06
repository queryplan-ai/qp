package mysql

import (
	"fmt"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	issuetypes "github.com/queryplan-ai/qp/pkg/issue/types"
)

func scanInsertStatementForIssues(query string, mysqlTables []MysqlTable) ([]issuetypes.QueryIssue, error) {
	_, err := parseInsertStatement(query)
	if err != nil {
		return nil, fmt.Errorf("parse insert statement: %w", err)
	}

	return nil, nil
}

func parseInsertStatement(query string) (*InsertStatement, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse insert statement: %w", err)
	}

	insertStmt, ok := stmt.(*sqlparser.Insert)
	if !ok {
		return nil, fmt.Errorf("parse insert statement: not an insert statement")
	}

	result := InsertStatement{
		Table:   "",
		Columns: []string{},
		Values:  [][]string{},
	}

	// Extract table name
	result.Table = insertStmt.Table.Name.String()

	// Extract column names
	for _, col := range insertStmt.Columns {
		result.Columns = append(result.Columns, col.String())
	}

	// Extract values - assuming simple cases for now
	// For handling more complex cases like subqueries, adjust accordingly
	if values, ok := insertStmt.Rows.(sqlparser.Values); ok {
		for _, valTuple := range values {
			var valueSet []string
			for _, val := range valTuple {
				valueSet = append(valueSet, sqlparser.String(val))
			}
			result.Values = append(result.Values, valueSet)
		}
	}

	return &result, nil
}

type InsertStatement struct {
	Table   string
	Columns []string
	Values  [][]string // Each slice within this slice represents a row of values
}
