package db

import (
	"database/sql"
	_ "github.com/lib/pq" // needed as the database is of type Postgres
	"github.com/nclandrei/L5-Project/jira"
)

const (
	connection   = "user=nclandrei password=nclandrei dbname=nclandrei sslmode=disable"
	databaseType = "postgres"
)

// JiraDatabase defines the jira database
type JiraDatabase struct {
	*sql.DB
}

// NewJiraDatabase returns a new database for the retrieved Jira issues
func NewJiraDatabase() (*JiraDatabase, error) {
	db, err := sql.Open(databaseType, connection)
	if err != nil {
		return nil, err
	}
	return &JiraDatabase{
		DB: db,
	}, nil
}

// GetIssues reads from the issues database and retrieves the issues as bytes
func (db *JiraDatabase) GetIssues() ([]jira.Issue, error) {
	rows, err := db.Query("SELECT * FROM ISSUES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var issues []jira.Issue
	for i := 0; rows.Next(); i++ {
		issue := new(jira.Issue)
		err = rows.Scan(
			&issue.Key,
			&issue.Fields.Summary,
			&issue.Fields.Description,
			&issue.Fields.TimeSpent,
			&issue.Fields.TimeEstimate,
			&issue.Fields.DueDate)
		if err != nil {
			return nil, err
		}
		issues = append(issues, *issue)
	}
	return issues, nil
}

// AddIssues inserts a slice of issues into the issues table
func (db *JiraDatabase) AddIssues(issues []jira.Issue) error {
	for _, issue := range issues {
		_, err := db.Exec("INSERT INTO ISSUE(KEY, SUMMARY, DESCRIPTION, TIME_SPENT, TIME_ESTIMATE, DUE_DATE) VALUES (?, ?, ?, ?, ?, ?);",
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.TimeSpent,
			issue.Fields.TimeEstimate,
			issue.Fields.DueDate)
		if err != nil {
			return err
		}
	}
	return nil
}
