package mysql

import (
	"fmt"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	dbtypes "github.com/queryplan-ai/qp/pkg/db/types"
	issuetypes "github.com/queryplan-ai/qp/pkg/issue/types"
	"github.com/queryplan-ai/qp/pkg/plan"
)

func PlanQuery(db *dbtypes.DB, query string) (string, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return "", err
	}

	mysqlTables := []MysqlTable{}
	for _, table := range db.Tables {
		mysqlTables = append(mysqlTables, table.(MysqlTable))
	}

	switch stmt.(type) {
	case *sqlparser.Select:
		issues, err := plan.ScanSelectStatementForIssues(query, db.Tables)
		if err != nil {
			return "", fmt.Errorf("scan select statement for issues: %w", err)
		}

		if len(issues) > 0 {
			return formatIssues(issues), nil
		}

		return "No issues found", nil
	case *sqlparser.Update:
		issues, err := scanUpdateStatementForIssues(query, mysqlTables)
		if err != nil {
			return "", fmt.Errorf("scan update statement for issues: %w", err)
		}

		if len(issues) > 0 {
			return formatIssues(issues), nil
		}

		return "No issues found", nil
	case *sqlparser.Insert:
		issues, err := scanInsertStatementForIssues(query, mysqlTables)
		if err != nil {
			return "", fmt.Errorf("scan insert statement for issues: %w", err)
		}

		if len(issues) > 0 {
			return formatIssues(issues), nil
		}

		return "No issues found", nil
	case *sqlparser.Delete:
		issues, err := scanDeleteStatementForIssues(query, mysqlTables)
		if err != nil {
			return "", fmt.Errorf("scan delete statement for issues: %w", err)
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
