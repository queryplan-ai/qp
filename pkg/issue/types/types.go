package types

import "time"

type QueryIssue struct {
	ID            string
	QueryID       string
	IssueSeverity string
	IssueType     string
	Message       string
	Data          string
	CreatedAt     time.Time
	IgnoredAt     *time.Time
	ResolvedAt    *time.Time
}

const (
	IssueSeverityLow    = "low"
	IssueSeverityMedium = "medium"
	IssueSeverityHigh   = "high"
)

const (
	TableIssueMissingPrimaryKey = "missing_primary_key"
)

const (
	QueryIssueTypeWhereClauseMissingIndex = "where_clause_missing_index"
	QueryIssueTypeJoinClauseMissingIndex  = "join_clause_missing_index"
	QueryIssueTypeColumnUpdatedInIndex    = "column_updated_in_index"
	QueryIssueTypeClauseMissingIndex      = "clause_missing_index"
)
