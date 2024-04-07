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
		postgresTable := tables[i].(PostgresTable)

		columns, err := listColumns(db, table.GetName())
		if err != nil {
			return nil, err
		}
		postgresTable.Columns = columns

		primaryKeys, err := listPrimaryKeys(db, table.GetName())
		if err != nil {
			return nil, err
		}
		postgresTable.PrimaryKeys = primaryKeys

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

func listPrimaryKeys(db *dbtypes.DB, tableName string) ([]string, error) {
	conn, err := connect(db.ConnectionURI)
	if err != nil {
		return nil, err
	}

	query := `select tc.constraint_name, c.column_name
from information_schema.table_constraints tc
join information_schema.constraint_column_usage as ccu using (constraint_schema, constraint_name)
join information_schema.columns as c on c.table_schema = tc.constraint_schema
  and tc.table_name = c.table_name and ccu.column_name = c.column_name
where constraint_type = 'PRIMARY KEY' and tc.table_name = $1
order by c.ordinal_position`

	rows, err := conn.Query(context.Background(), query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query primary keys: %w", err)
	}
	defer rows.Close()

	primaryKeys := []string{}
	for rows.Next() {
		var constraintName, columnName string

		if err := rows.Scan(&constraintName, &columnName); err != nil {
			return nil, err
		}

		primaryKeys = append(primaryKeys, columnName)
	}
	return primaryKeys, nil
}

func indexesByTable(postgresTables []PostgresTable) map[string][]Index {
	indexesByTable := make(map[string][]Index)
	for _, postgresTable := range postgresTables {
		// primary keys
		indexesByTable[postgresTable.TableName] = append(indexesByTable[postgresTable.TableName], Index{
			Columns:      postgresTable.PrimaryKeys,
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
	}

	return indexesByTable
}
