package shell

import (
	"fmt"
	"net/url"

	"github.com/queryplan-ai/qp/pkg/db"
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
	"github.com/queryplan-ai/qp/pkg/shell/types"
)

var (
	ErrUnsupportedScheme = fmt.Errorf("unsupported connection scheme")
)

func handleConnect(sh *types.Shell, cmd string) *types.ShellCommandResult {
	result := &types.ShellCommandResult{
		IsFatal:   false,
		IsSuccess: false,
	}

	// parse the connection string
	uri, err := url.Parse(cmd)
	if err != nil {
		result.Message = fmt.Sprintf("Error parsing connection string: %s", err)
		return result
	}

	switch uri.Scheme {
	case "mysql":
		dbName, err := db.VerifyMysqlConnection(cmd)
		if err != nil {
			result.Message = fmt.Sprintf("Error connecting to database: %s", err)
			return result
		}

		sh.DatabaseName = dbName
		sh.DatabaseEngine = "mysql"

	case "postgres", "postgresql":
		// test the connection
		dbName, err := db.VerifyPGConnection(cmd)
		if err != nil {
			result.Message = fmt.Sprintf("Error connecting to database: %s", err)
			return result
		}

		sh.DatabaseName = dbName
		sh.DatabaseEngine = "postgres"

	default:
		result.Message = ErrUnsupportedScheme.Error()
		return result
	}

	sh.DB = &dbtypes.DB{
		ConnectionURI: cmd,
		DatabaseName:  sh.DatabaseName,
	}

	go db.LoadSchema(sh.DB)

	result.IsSuccess = true
	return result
}
