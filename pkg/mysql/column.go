package mysql

import (
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
)

var _ dbtypes.Column = MysqlColumn{}

type MysqlColumn struct {
	ColumnName    string
	DataType      string
	ColumnType    string
	IsNullable    bool
	ColumnKey     string
	ColumnDefault *string
	Extra         string
}

func (c MysqlColumn) GetName() string {
	return c.ColumnName
}

func (c MysqlColumn) GetDataType() string {
	return c.DataType
}

func (c MysqlColumn) GetColumnType() string {
	return c.ColumnType
}

func (c MysqlColumn) GetIsNullable() bool {
	return c.IsNullable
}

func (c MysqlColumn) GetColumnKey() string {
	return c.ColumnKey
}

func (c MysqlColumn) GetColumnDefault() *string {
	return c.ColumnDefault
}

func (c MysqlColumn) GetExtra() string {
	return c.Extra
}
