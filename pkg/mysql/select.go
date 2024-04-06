package mysql

import (
	"fmt"
	"strings"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	issuetypes "github.com/queryplan-ai/qp/pkg/issue/types"
)

type SelectStatement struct {
	Columns map[string][]string
	Tables  []string
	Where   map[string][]string
	Join    map[string][]string
}

type Index struct {
	Columns      []string
	IsPrimaryKey bool
	IsUnique     bool
}

func scanSelectStatementForIssues(query string, tables []MysqlTable) ([]issuetypes.QueryIssue, error) {
	selectStatement, err := parseSelectStatement(query, tables)
	if err != nil {
		return nil, err
	}

	if selectStatement == nil {
		return nil, nil
	}

	// build our index map
	indexesByTable := make(map[string][]Index)
	for _, table := range tables {
		// primary keys
		indexesByTable[table.TableName] = append(indexesByTable[table.TableName], Index{
			Columns:      table.PrimaryKeys,
			IsPrimaryKey: true,
			IsUnique:     true, // of course
		})

		// other indexes
		// for _, index := range table.Indexes {
		// 	indexesByTable[table.TableName] = append(indexesByTable[table.TableName], Index{
		// 		Columns:      index.Columns,
		// 		IsPrimaryKey: false,
		// 		IsUnique:     index.IsUnique,
		// 	})
		// }
	}

	issues, err := scanSelectStatementForMissingIndexes(query, selectStatement, tables, indexesByTable)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// parseSelectStatement will parse a select statement and return a SelectStatement struct
// that we use for further analysis.
func parseSelectStatement(query string, mysqlTables []MysqlTable) (*SelectStatement, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse select statement: %w", err)
	}

	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return nil, fmt.Errorf("expected select statement, got %T", stmt)
	}

	// Check if the query is from information_schema
	if selectStmt.From != nil {
		for _, tableExpr := range selectStmt.From {
			if aliasedTableExpr, ok := tableExpr.(*sqlparser.AliasedTableExpr); ok {
				switch expr := aliasedTableExpr.Expr.(type) {
				case sqlparser.TableName:
					qualifier := strings.ToLower(expr.Qualifier.String())
					if qualifier == "information_schema" {
						return nil, nil
					}
				case *sqlparser.Subquery:
					// parse this subquery
				default:
					fmt.Printf("Unexpected type: %T\n", expr)
				}
			}
		}
	}

	result := SelectStatement{
		Columns: map[string][]string{},
		Tables:  []string{},
		Where:   map[string][]string{},
		Join:    map[string][]string{},
	}

	tableAliasLookup, tables, err := extractTables(selectStmt)
	if err != nil {
		return nil, fmt.Errorf("extract tables: %w", err)
	}

	result.Tables = tables

	err = processSelectExpressions(selectStmt, tableAliasLookup, mysqlTables, &result)
	if err != nil {
		return nil, fmt.Errorf("process select expressions: %w", err)
	}

	if err := processJoinClauses(selectStmt.From, tableAliasLookup, &result); err != nil {
		return nil, fmt.Errorf("process join clauses: %w", err)
	}

	if selectStmt.Where != nil {
		if err := processWhereClause(selectStmt.Where.Expr, tableAliasLookup, mysqlTables, &result); err != nil {
			return nil, fmt.Errorf("process where clause: %w", err)
		}
	}

	return &result, nil
}

func scanSelectStatementForMissingIndexes(cleanedStatement string, selectStatement *SelectStatement, tables []MysqlTable, indexesByTable map[string][]Index) ([]issuetypes.QueryIssue, error) {
	queryIssues := []issuetypes.QueryIssue{}

	// check if the where clause contains a column that is not indexed
	for table, columns := range selectStatement.Where {
		if _, exists := indexesByTable[table]; !exists {
			continue
		}

		for _, index := range indexesByTable[table] {
			for _, column := range columns {
				if !contains(index.Columns, column) {
					columnNames := ""
					columnNames += fmt.Sprintf(`From the %q" table: %s`, table, strings.Join(selectStatement.Columns[table], ", "))

					queryIssues = append(queryIssues, issuetypes.QueryIssue{
						IssueSeverity: issuetypes.IssueSeverityLow,
						IssueType:     issuetypes.QueryIssueTypeWhereClauseMissingIndex,
						Message:       "where clause contains a column that is not indexed",
					})
				}
			}
		}
	}

	// check if the join clause contains a column that is not indexed
	for table, columns := range selectStatement.Join {
		if _, exists := indexesByTable[table]; !exists {
			continue
		}

		for _, index := range indexesByTable[table] {
			for _, column := range columns {
				if !contains(index.Columns, column) {
					columnNames := ""
					columnNames += fmt.Sprintf(`From the %q" table: %s`, table, strings.Join(selectStatement.Columns[table], ", "))

					queryIssues = append(queryIssues, issuetypes.QueryIssue{
						IssueSeverity: issuetypes.IssueSeverityLow,
						IssueType:     issuetypes.QueryIssueTypeClauseMissingIndex,
						Message:       "join clause contains a column that is not indexed",
					})
				}
			}
		}
	}

	return queryIssues, nil
}

func extractTables(selectStmt *sqlparser.Select) (map[string]string, []string, error) {
	tableAliasLookup := make(map[string]string)
	tables := make([]string, 0)

	var extractFunc func(node sqlparser.SQLNode) (kontinue bool, err error)
	extractFunc = func(node sqlparser.SQLNode) (bool, error) {
		switch node := node.(type) {
		case *sqlparser.AliasedTableExpr:
			var fullTableName string
			if tbl, ok := node.Expr.(*sqlparser.TableName); ok {
				// Handles fully qualified table names
				fullTableName = tbl.Name.String()
				if tbl.Qualifier.String() != "" {
					fullTableName = tbl.Qualifier.String() + "." + fullTableName
				}
			} else {
				// Fallback for other expressions, if necessary
				fullTableName = sqlparser.String(node.Expr)
			}

			alias := sqlparser.String(node.As)
			if alias == "" {
				alias = fullTableName
			}
			tableAliasLookup[alias] = fullTableName

			if !contains(tables, fullTableName) {
				tables = append(tables, fullTableName)
			}

		case *sqlparser.JoinTableExpr:
			// Recursively extract tables from the left and right expressions of the join
			if _, err := extractFunc(node.LeftExpr); err != nil {
				return false, err
			}
			if _, err := extractFunc(node.RightExpr); err != nil {
				return false, err
			}
		}
		return true, nil
	}

	if err := sqlparser.Walk(extractFunc, selectStmt.From); err != nil {
		return nil, nil, fmt.Errorf("walk: %w", err)
	}

	return tableAliasLookup, tables, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func columnNamesForTable(table string, mysqlTables []MysqlTable) []string {
	for _, t := range mysqlTables {
		if t.TableName == table {
			columnNames := []string{}
			for _, col := range t.Columns {
				columnNames = append(columnNames, col.ColumnName)
			}

			return columnNames
		}
	}

	return nil
}

func processSelectExpressions(selectStmt *sqlparser.Select, tableAliasLookup map[string]string, mysqlTables []MysqlTable, result *SelectStatement) error {
	for _, selectExpr := range selectStmt.SelectExprs {
		switch expr := selectExpr.(type) {
		case *sqlparser.StarExpr:
			// Handle wildcard cases
			if tableName, err := resolveTableName(expr.TableName, tableAliasLookup, len(result.Tables)); err == nil && tableName != "" {
				result.Columns[tableName] = append(result.Columns[tableName], columnNamesForTable(tableName, mysqlTables)...)
			}

		case *sqlparser.AliasedExpr:
			// Handle column expressions
			if col, ok := expr.Expr.(*sqlparser.ColName); ok {
				qualifier := col.Qualifier.Name.String()
				column := col.Name.String()
				if tableName, err := resolveColumnTable(result.Tables, qualifier, column, tableAliasLookup, mysqlTables); err == nil && tableName != "" {
					result.Columns[tableName] = append(result.Columns[tableName], column)
				}
			} else if funcExpr, ok := expr.Expr.(*sqlparser.FuncExpr); ok {
				// Handle function expressions
				funcName := strings.ToUpper(funcExpr.Name.String())
				if len(result.Tables) == 1 {
					// If there's only one table, associate the function with that table
					tableName := result.Tables[0]
					result.Columns[tableName] = append(result.Columns[tableName], funcName)
				}
				// Additional logic can be added here for more complex scenarios
			}
		}
	}

	if selectStmt.Where != nil {
		if err := processWhereClause(selectStmt.Where.Expr, tableAliasLookup, mysqlTables, result); err != nil {
			return fmt.Errorf("process where clause: %w", err)
		}
	}
	return nil
}
func resolveTableName(tableName sqlparser.TableName, tableAliasLookup map[string]string, numTables int) (string, error) {
	if !tableName.IsEmpty() {
		alias := tableName.Name.String()
		if actualName, exists := tableAliasLookup[alias]; exists {
			return actualName, nil
		}
		return tableName.Name.String(), nil
	} else if numTables == 1 {
		// If no table name is provided with the wildcard, and there is only one table
		for _, actualName := range tableAliasLookup {
			return actualName, nil // Return the single table name
		}
	}
	return "", fmt.Errorf("table name %q not found", tableName.Name.String())
}

func resolveColumnTable(tables []string, qualifier, column string, tableAliasLookup map[string]string, mysqlTables []MysqlTable) (string, error) {
	fullQualifier := qualifier
	if qualifier != "" {
		if actualTable, exists := tableAliasLookup[fullQualifier]; exists {
			return actualTable, nil
		}
		return "", fmt.Errorf("table alias %q not found", fullQualifier)
	} else {
		// only search columns by table that also exist in tables array
		for _, table := range tables {
			if contains(columnNamesForTable(table, mysqlTables), column) {
				return table, nil
			}
		}
	}

	return "", fmt.Errorf("column %q not found in any table", column)
}

func sliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func processJoinClauses(tableExprs sqlparser.TableExprs, tableAliasLookup map[string]string, result *SelectStatement) error {
	for _, tableExpr := range tableExprs {
		switch expr := tableExpr.(type) {
		case *sqlparser.JoinTableExpr:
			// Process the join expression
			if err := extractJoinColumns(expr.On, tableAliasLookup, result); err != nil {
				return err
			}

			// Recursively process the left and right sides of the join
			if err := processJoinClauses([]sqlparser.TableExpr{expr.LeftExpr}, tableAliasLookup, result); err != nil {
				return err
			}
			if err := processJoinClauses([]sqlparser.TableExpr{expr.RightExpr}, tableAliasLookup, result); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractJoinColumns(onExpr sqlparser.Expr, tableAliasLookup map[string]string, result *SelectStatement) error {
	switch expr := onExpr.(type) {
	case *sqlparser.ComparisonExpr:
		// Handle comparison expressions
		if err := extractColumnFromExpr(expr.Left, tableAliasLookup, result); err != nil {
			return err
		}
		if err := extractColumnFromExpr(expr.Right, tableAliasLookup, result); err != nil {
			return err
		}
		// Handle other cases as needed
	}
	return nil
}

func extractColumnFromExpr(expr sqlparser.Expr, tableAliasLookup map[string]string, result *SelectStatement) error {
	if colExpr, ok := expr.(*sqlparser.ColName); ok {
		// Use the Qualifier directly as it is already a sqlparser.TableName
		tableName := colExpr.Qualifier

		resolvedTableName, err := resolveTableName(tableName, tableAliasLookup, len(result.Tables))
		if err != nil {
			return err
		}

		columnName := colExpr.Name.String()
		result.Join[resolvedTableName] = appendIfMissing(result.Join[resolvedTableName], columnName)
	}
	return nil
}

func appendIfMissing(slice []string, element string) []string {
	for _, elem := range slice {
		if elem == element {
			return slice // Element already present, no need to append
		}
	}
	return append(slice, element) // Element not present, append it
}

func processWhereClause(whereExpr sqlparser.Expr, tableAliasLookup map[string]string, mysqlTables []MysqlTable, result *SelectStatement) error {
	// This function will need to recursively process the WHERE clause expression tree
	// and extract column names used in the expressions.
	switch expr := whereExpr.(type) {
	case *sqlparser.ColName:
		qualifier := expr.Qualifier.Name.String()
		column := expr.Name.String()
		tableName, err := resolveColumnTable(result.Tables, qualifier, column, tableAliasLookup, mysqlTables)
		if err != nil {
			return fmt.Errorf("resolve column table: %w", err)
		}
		if tableName != "" {
			// Check if column already exists for the table
			if !sliceContains(result.Where[tableName], column) {
				result.Where[tableName] = append(result.Where[tableName], column)
			}
		}
	case *sqlparser.ComparisonExpr:
		// Handle comparison expressions
		if err := processWhereClause(expr.Left, tableAliasLookup, mysqlTables, result); err != nil {
			return fmt.Errorf("process where clause (left): %w", err)
		}
		if err := processWhereClause(expr.Right, tableAliasLookup, mysqlTables, result); err != nil {
			return fmt.Errorf("process where clause (right): %w", err)
		}

	case *sqlparser.AndExpr, *sqlparser.OrExpr:
		// Handle AND, OR expressions (they have the same structure)
		if binaryExpr, ok := expr.(*sqlparser.BinaryExpr); ok {
			if err := processWhereClause(binaryExpr.Left, tableAliasLookup, mysqlTables, result); err != nil {
				return fmt.Errorf("process where clause (and/or left): %w", err)
			}
			if err := processWhereClause(binaryExpr.Right, tableAliasLookup, mysqlTables, result); err != nil {
				return fmt.Errorf("process where clause (and/or right): %w", err)
			}
		}

	case *sqlparser.ParenExpr:
		// Handle parenthesized expressions
		if err := processWhereClause(expr.Expr, tableAliasLookup, mysqlTables, result); err != nil {
			return fmt.Errorf("process where clause (paren): %w", err)
		}

	default:
		// Handle other types of expressions (subqueries, functions, etc.)
	}

	return nil
}
