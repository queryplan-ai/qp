package mysql

import (
	"database/sql"
	"fmt"

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
		return err
	}

	primaryKeys, err := listPrimaryKeys(db)
	if err != nil {
		return err
	}

	for i, table := range tables {
		if _, ok := primaryKeys[table.GetName()]; !ok {
			primaryKeys[table.GetName()] = []string{}
		}

		mysqlTable := tables[i].(MysqlTable)
		mysqlTable.PrimaryKeys = primaryKeys[table.GetName()]
		tables[i] = mysqlTable
	}

	db.SchemaLoaded = true
	db.Tables = tables

	return nil
}

func listPrimaryKeys(db *dbtypes.DB) (map[string][]string, error) {
	conn, err := connect(db.ConnectionURI)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query("SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME FROM  INFORMATION_SCHEMA.KEY_COLUMN_USAGE  WHERE  CONSTRAINT_NAME = 'PRIMARY' AND TABLE_SCHEMA = ? ORDER BY TABLE_NAME, ORDINAL_POSITION", db.DatabaseName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	primaryKeys := map[string][]string{}
	for rows.Next() {
		tableName := ""
		columnName := ""
		if err := rows.Scan(&tableName, &tableName, &columnName); err != nil {
			return nil, fmt.Errorf("scan primary keys: %w", err)
		}

		if _, ok := primaryKeys[tableName]; !ok {
			primaryKeys[tableName] = []string{}
		}

		primaryKeys[tableName] = append(primaryKeys[tableName], columnName)
	}

	return primaryKeys, nil
}

func listTables(db *dbtypes.DB) ([]dbtypes.Table, error) {
	// read the schema from mysql
	conn, err := connect(db.ConnectionURI)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(`SELECT
c.TABLE_NAME, c.COLUMN_NAME, c.DATA_TYPE, c.COLUMN_TYPE, c.IS_NULLABLE, c.COLUMN_KEY, c.COLUMN_DEFAULT, c.EXTRA,
t.TABLE_ROWS
FROM INFORMATION_SCHEMA.COLUMNS c
INNER JOIN INFORMATION_SCHEMA.TABLES t ON t.TABLE_NAME = c.TABLE_NAME AND t.TABLE_SCHEMA = c.TABLE_SCHEMA
WHERE c.TABLE_SCHEMA = ?`, db.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}

	defer rows.Close()

	tables := []dbtypes.Table{}
	for rows.Next() {
		column := MysqlColumn{}

		tableName := ""
		estimatedRowCount := int64(0)
		isNullable := ""
		columnDefault := sql.NullString{}
		if err := rows.Scan(&tableName, &column.ColumnName, &column.DataType, &column.ColumnType, &isNullable, &column.ColumnKey, &columnDefault, &column.Extra, &estimatedRowCount); err != nil {
			return nil, err
		}

		if isNullable == "YES" {
			column.IsNullable = true
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		found := false
		for i, table := range tables {
			if table.GetName() == tableName {
				existingTable := tables[i].(MysqlTable)
				existingTable.Columns = append(existingTable.Columns, column)
				tables[i] = existingTable
				found = true
				continue
			}
		}

		if !found {
			mysqlTable := MysqlTable{
				TableName:         tableName,
				Columns:           []MysqlColumn{column},
				EstimatedRowCount: estimatedRowCount,
			}

			tables = append(tables, mysqlTable)
		}
	}

	return tables, nil
}

type Index struct {
	Columns      []string
	IsPrimaryKey bool
	IsUnique     bool
}

func indexesByTable(mysqlTables []MysqlTable) map[string][]Index {
	indexesByTable := make(map[string][]Index)
	for _, mysqlTable := range mysqlTables {
		// primary keys
		indexesByTable[mysqlTable.TableName] = append(indexesByTable[mysqlTable.TableName], Index{
			Columns:      mysqlTable.PrimaryKeys,
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
