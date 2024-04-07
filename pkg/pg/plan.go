package pg

import (
	"fmt"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
	issuetypes "github.com/queryplan-ai/qp/pkg/issue/types"
)

func PlanQuery(db *dbtypes.DB, query string) (string, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return "", err
	}

	postgresTables := []PostgresTable{}
	for _, table := range db.Tables {
		postgresTables = append(postgresTables, table.(PostgresTable))
	}

	switch stmt.(type) {
	case *sqlparser.Select:
		issues, err := scanSelectStatementForIssues(query, postgresTables)
		if err != nil {
			return "", fmt.Errorf("scan select statement for issues: %w", err)
		}

		if len(issues) > 0 {
			return formatIssues(issues), nil
		}

		return "No issues found", nil
	}

	return "", nil
}

func formatIssues(issues []issuetypes.QueryIssue) string {
	var formattedIssues string
	for _, issue := range issues {
		formattedIssues += issue.Message + "\n"
	}

	return formattedIssues
}
