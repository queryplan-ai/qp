package types

import (
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
)

type ShellOpts struct {
	ConnectionURI string
	OpenAIAPIKey  string
}

type Shell struct {
	DB *dbtypes.DB

	DatabaseName   string
	DatabaseEngine string

	HistoryFilePath string
	HistoryMaxSize  int
}

type ShellCommandResult struct {
	IsSuccess bool
	IsFatal   bool
	Message   string
}
