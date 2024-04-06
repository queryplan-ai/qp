package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseSelectStatement(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		mysqlTables []MysqlTable
		want        *SelectStatement
	}{
		{
			name:  "simple select, one column, one table",
			query: "SELECT id FROM users",
			mysqlTables: []MysqlTable{
				{
					TableName: "users",
					Columns: []MysqlColumn{
						{
							ColumnName: "id",
							DataType:   "int",
						},
					},
				},
			},
			want: &SelectStatement{
				Columns: map[string][]string{
					"users": {"id"},
				},
				Tables: []string{"users"},
				Where:  map[string][]string{},
				Join:   map[string][]string{},
			},
		},
		{
			name:  "one column, two tables, inner join",
			query: `SELECT users.id FROM users INNER JOIN orders ON users.id = orders.user_id WHERE users.id = 1`,
			mysqlTables: []MysqlTable{
				{
					TableName: "users",
					Columns: []MysqlColumn{
						{
							ColumnName: "id",
							DataType:   "int",
						},
					},
				},
				{
					TableName: "orders",
					Columns: []MysqlColumn{
						{
							ColumnName: "id",
							DataType:   "int",
						},
						{
							ColumnName: "user_id",
							DataType:   "int",
						},
					},
				},
			},
			want: &SelectStatement{
				Columns: map[string][]string{
					"users": {"id"},
				},
				Tables: []string{"users", "orders"},
				Where: map[string][]string{
					"users": {"id"},
				},
				Join: map[string][]string{
					"users":  {"id"},
					"orders": {"user_id"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			got, err := parseSelectStatement(tt.query, tt.mysqlTables)
			req.NoError(err)

			assert.Equal(t, tt.want, got)
		})
	}
}
