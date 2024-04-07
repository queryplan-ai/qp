package pg

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
)

func LoadSchema(db *dbtypes.DB) error {
	db.SchemaLoading = true
	db.SchemaLoaded = false

	defer func() {
		db.SchemaLoading = false
	}()

	tables, err := listTables(db)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}

	db.SchemaLoaded = true
	db.Tables = tables

	fmt.Printf("Loaded tables : %v\n", tables)

	return nil
}

func listTables(db *dbtypes.DB) ([]dbtypes.Table, error) {
	conn, err := connect(db.ConnectionURI)
	if err != nil {
		return nil, err
	}

	query := "select table_name from information_schema.tables where table_catalog = $1 and table_schema = $2"

	rows, err := conn.Query(context.Background(), query, db.DatabaseName, "public")
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	tables := []dbtypes.Table{}
	for rows.Next() {
		tableName := ""
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan tables: %w", err)
		}

		postgresTable := PostgresTable{
			TableName: tableName,
		}

		tables = append(tables, postgresTable)
	}

	// load columns for each table
	for i, table := range tables {
		columns, err := listColumns(db, table.GetName())
		if err != nil {
			return nil, err
		}

		postgresTable := tables[i].(PostgresTable)
		postgresTable.Columns = columns
		tables[i] = postgresTable
	}

	return tables, nil
}

func listColumns(db *dbtypes.DB, tableName string) ([]PostgresColumn, error) {
	conn, err := connect(db.ConnectionURI)
	if err != nil {
		return nil, err
	}

	query := "select column_name, data_type, character_maximum_length, column_default, is_nullable from information_schema.columns where table_name = $1 and table_catalog = $2"

	rows, err := conn.Query(context.Background(), query, tableName, db.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	columns := []PostgresColumn{}
	for rows.Next() {
		postgresColumn := PostgresColumn{}

		var maxLength sql.NullInt64
		var isNullable string
		var columnDefault sql.NullString

		if err := rows.Scan(&postgresColumn.ColumnName, &postgresColumn.DataType, &maxLength, &columnDefault, &isNullable); err != nil {
			return nil, fmt.Errorf("scan columns: %w", err)
		}

		postgresColumn.IsNullable = isNullable == "YES"

		if columnDefault.Valid {
			value := stripOIDClass(columnDefault.String)
			postgresColumn.ColumnDefault = &value
		}

		if maxLength.Valid {
			postgresColumn.DataType = fmt.Sprintf("%s (%d)", postgresColumn.DataType, maxLength.Int64)
		}

		columns = append(columns, postgresColumn)
	}

	return columns, nil
}

var oidClassRegexp = regexp.MustCompile(`'(.*)'::.+`)

func stripOIDClass(value string) string {
	matches := oidClassRegexp.FindStringSubmatch(value)
	if len(matches) == 2 {
		return matches[1]
	}
	return value
}
