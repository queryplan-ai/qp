package mysql

import (
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
)

var _ dbtypes.Table = MysqlTable{}

type MysqlTable struct {
	TableName         string
	Columns           []MysqlColumn
	PrimaryKeys       []string
	EstimatedRowCount int64
}

func (t MysqlTable) GetName() string {
	return t.TableName
}

func (t MysqlTable) GetColumns() []dbtypes.Column {
	var cols []dbtypes.Column
	for _, c := range t.Columns {
		cols = append(cols, c)
	}
	return cols
}

func (t MysqlTable) GetPrimaryKeys() []string {
	return t.PrimaryKeys
}

func (t MysqlTable) GetEstimatedRowCount() int64 {
	return t.EstimatedRowCount
}
