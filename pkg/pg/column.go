package pg

import (
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
)

var _ dbtypes.Column = PostgresColumn{}

type PostgresColumn struct {
	ColumnName    string
	DataType      string
	ColumnType    string
	IsNullable    bool
	ColumnKey     string
	ColumnDefault *string
	Extra         string
}

func (c PostgresColumn) GetName() string {
	return c.ColumnName
}

func (c PostgresColumn) GetDataType() string {
	return c.DataType
}

func (c PostgresColumn) GetColumnType() string {
	return c.ColumnType
}

func (c PostgresColumn) GetIsNullable() bool {
	return c.IsNullable
}

func (c PostgresColumn) GetColumnKey() string {
	return c.ColumnKey
}

func (c PostgresColumn) GetColumnDefault() *string {
	return c.ColumnDefault
}

func (c PostgresColumn) GetExtra() string {
	return c.Extra
}
