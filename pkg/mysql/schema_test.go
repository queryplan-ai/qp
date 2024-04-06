package mysql

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_indexesByTable(t *testing.T) {
	tests := []struct {
		name        string
		mysqlTables []MysqlTable
		want        map[string][]Index
	}{
		{
			name: "one primary key",
			mysqlTables: []MysqlTable{
				{
					TableName: "table1",
					Columns: []MysqlColumn{
						{
							ColumnName:    "id",
							DataType:      "int",
							ColumnType:    "int(11)",
							IsNullable:    false,
							ColumnKey:     "PRI",
							ColumnDefault: nil,
							Extra:         "",
						},
					},
					PrimaryKeys:       []string{"id"},
					EstimatedRowCount: 0,
				},
			},
			want: map[string][]Index{
				"table1": {
					{
						Columns:      []string{"id"},
						IsPrimaryKey: true,
						IsUnique:     true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexesByTable(tt.mysqlTables)
			assert.True(t, reflect.DeepEqual(got, tt.want))
		})
	}
}
