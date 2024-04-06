package types

type DB struct {
	ConnectionURI string
	DatabaseName  string

	SchemaLoading bool
	SchemaLoaded  bool

	Tables []Table
}

type Table interface {
	GetName() string
	GetColumns() []Column
	GetPrimaryKeys() []string
	GetEstimatedRowCount() int64
}

type Column interface {
	GetName() string
	GetDataType() string
	GetColumnType() string
	GetIsNullable() bool
	GetColumnKey() string
	GetColumnDefault() *string
	GetExtra() string
}
