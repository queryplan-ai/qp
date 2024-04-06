package mysql

import (
	"fmt"
	"strings"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	issuetypes "github.com/queryplan-ai/qp/pkg/issue/types"
)

type UpdateStatement struct {
	Columns map[string][]string
	Tables  []string
}

func scanUpdateStatementForIssues(query string, mysqlTables []MysqlTable) ([]issuetypes.QueryIssue, error) {
	updateStatement, err := parseUpdateStatement(query, mysqlTables)
	if err != nil {
		return nil, fmt.Errorf("parse update statement: %w", err)
	}

	if updateStatement == nil {
		return nil, nil
	}

	queryIssues := []issuetypes.QueryIssue{}

	issues, err := scanUpdateStatementForMissingIndexes(updateStatement, indexesByTable(mysqlTables))
	if err != nil {
		return nil, err
	}
	queryIssues = append(queryIssues, issues...)

	issues, err = scanUpdateStatementForIndexUpdates(query, mysqlTables, updateStatement, indexesByTable(mysqlTables))
	if err != nil {
		return nil, err
	}
	queryIssues = append(queryIssues, issues...)

	return queryIssues, nil
}

func scanUpdateStatementForIndexUpdates(query string, mysqlTables []MysqlTable, updateStatement *UpdateStatement, indexesByTable map[string][]Index) ([]issuetypes.QueryIssue, error) {
	queryIssues := []issuetypes.QueryIssue{}

	// check if the column that's updated is part of an index
	for table, columns := range updateStatement.Columns {
		if _, exists := indexesByTable[table]; !exists {
			continue
		}

		for _, index := range indexesByTable[table] {
			for _, column := range columns {
				if contains(index.Columns, column) {
					columnNames := ""
					columnNames += fmt.Sprintf(`From the %q" table: %s`, table, strings.Join(updateStatement.Columns[table], ", "))

					queryIssues = append(queryIssues, issuetypes.QueryIssue{
						IssueSeverity: issuetypes.IssueSeverityLow,
						IssueType:     issuetypes.QueryIssueTypeColumnUpdatedInIndex,
						Message:       "column updated is part of an index",
					})
				}
			}
		}
	}

	return queryIssues, nil
}

func parseUpdateStatement(query string, mysqlTables []MysqlTable) (*UpdateStatement, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse query: %w", err)
	}

	updateStmt, ok := stmt.(*sqlparser.Update)
	if !ok {
		return nil, fmt.Errorf("not an update statement")
	}

	result := UpdateStatement{
		Columns: make(map[string][]string),
		Tables:  []string{},
	}

	// Extract table names
	result.Tables = extractUpdateTableName(updateStmt)

	// Extract columns and their new values
	err = processUpdateExpressions(updateStmt, mysqlTables, result.Tables, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func extractUpdateTableName(updateStmt *sqlparser.Update) []string {
	var tables []string
	for _, tableExpr := range updateStmt.TableExprs {
		switch expr := tableExpr.(type) {
		case *sqlparser.AliasedTableExpr:
			tableName := sqlparser.String(expr.Expr)
			tables = append(tables, tableName)
		case *sqlparser.JoinTableExpr:
			// Handle JOINs
			leftTable := sqlparser.String(expr.LeftExpr)
			rightTable := sqlparser.String(expr.RightExpr)
			tables = append(tables, leftTable, rightTable)
			// Note: This is a simplified handling. For more complex JOINs, further parsing may be required.
		}
	}
	return tables
}

func processUpdateExpressions(updateStmt *sqlparser.Update, mysqlTables []MysqlTable, tables []string, result *UpdateStatement) error {
	for _, updateExpr := range updateStmt.Exprs {
		columnName := sqlparser.String(updateExpr.Name.Name)

		// Check if the column belongs to any of the tables being updated
		for _, table := range tables {
			if contains(columnNamesForTable(table, mysqlTables), columnName) {
				// Add the column to the respective table's slice in the result
				result.Columns[table] = append(result.Columns[table], columnName)
				break
			} else {
				// If the table is not in columnsByTable, assume it's valid and add it
				result.Columns[table] = append(result.Columns[table], columnName)
				break
			}
		}
	}
	return nil
}

func scanUpdateStatementForMissingIndexes(updateStatement *UpdateStatement, indexesByTable map[string][]Index) ([]issuetypes.QueryIssue, error) {
	queryIssues := []issuetypes.QueryIssue{}

	return queryIssues, nil
}
