package db

import (
	"database/sql"
	"fmt"
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
	errs := ""
	for _, issue := range issues {
		_, err := db.Exec("INSERT INTO issue VALUES ($1, $2, $3, $4, $5, $6);",
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.TimeSpent,
			issue.Fields.TimeEstimate,
			issue.Fields.DueDate)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}
