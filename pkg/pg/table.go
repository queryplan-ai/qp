package pg

import (
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
)

var _ dbtypes.Table = PostgresTable{}

type PostgresTable struct {
	TableName         string
	Columns           []PostgresColumn
	PrimaryKeys       []string
	EstimatedRowCount int64
}

func (t PostgresTable) GetName() string {
	return t.TableName
}

func (t PostgresTable) GetColumns() []dbtypes.Column {
	var cols []dbtypes.Column
	for _, c := range t.Columns {
		cols = append(cols, c)
	}
	return cols
}

func (t PostgresTable) GetPrimaryKeys() []string {
	return t.PrimaryKeys
}

func (t PostgresTable) GetEstimatedRowCount() int64 {
	return t.EstimatedRowCount
}
