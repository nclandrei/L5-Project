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
	db.SetMaxOpenConns(100000)
	db.SetMaxIdleConns(100000)
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

// InsertIssues inserts a slice of issues into the issues table
func (db *JiraDatabase) InsertIssues(issues []jira.Issue) error {
	var errs string
	// issueCh := make(chan string, len(issues))
	for _, issue := range issues {
		// go func(issue jira.Issue, issueCh chan string) {
		errs := ""
		_, err := db.Exec("INSERT INTO issue VALUES ($1, $2, $3, $4, $5, $6);",
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.TimeSpent,
			issue.Fields.TimeEstimate,
			issue.Fields.DueDate)
		if err != nil {
			errs += fmt.Sprintf("Could not insert issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertPriority(db, issue.Key, issue.Fields.Priority)
		if err != nil {
			errs += fmt.Sprintf("Could not insert priority for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertIssueType(db, issue.Key, issue.Fields.IssueType)
		if err != nil {
			errs += fmt.Sprintf("Could not insert issue type for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertComments(db, issue.Key, issue.Fields.Comment)
		if err != nil {
			errs += fmt.Sprintf("Could not insert comments for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertStatus(db, issue.Key, issue.Fields.Status)
		if err != nil {
			errs += fmt.Sprintf("Could not insert status for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertChangelog(db, issue.Key, issue.Changelog)
		if err != nil {
			errs += fmt.Sprintf("Could not insert changelog for issue %s: %s\n", issue.Key, err.Error())
		}
		// issueCh <- errs
		// }(issue, issueCh)
	}
	// for i := 0; i < len(issues); i++ {
	// 	errs += <-issueCh
	// }
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}

func insertComments(db *JiraDatabase, issueKey string, comments []jira.Comment) error {
	errs := ""
	for _, comment := range comments {
		_, err := db.Exec("INSERT INTO comment VALUES ($1, $2, $3, $4, $5);",
			comment.ID,
			issueKey,
			comment.Body,
			comment.Created,
			comment.Updated,
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		_, err = db.Exec("INSERT INTO comment_author VALUES ($1, $2);",
			comment.ID,
			comment.Author.Name,
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}

func insertPriority(db *JiraDatabase, issueKey string, priority jira.Priority) error {
	_, err := db.Exec("INSERT INTO priority VALUES ($1, $2, $3);",
		issueKey,
		priority.ID,
		priority.Name,
	)
	if err != nil {
		return err
	}
	return nil
}

func insertIssueType(db *JiraDatabase, issueKey string, issueType jira.IssueType) error {
	_, err := db.Exec("INSERT INTO issue_type VALUES ($1, $2, $3, $4);",
		issueKey,
		issueType.ID,
		issueType.Name,
		issueType.Description,
	)
	if err != nil {
		return err
	}
	return nil
}

func insertStatus(db *JiraDatabase, issueKey string, status jira.Status) error {
	_, err := db.Exec("INSERT INTO issue_type VALUES ($1, $2, $3, $4);",
		issueKey,
		status.Description,
		status.ID,
		status.Name,
	)
	if err != nil {
		return err
	}
	return nil
}

func insertChangelog(db *JiraDatabase, issueKey string, changelog jira.Changelog) error {
	errs := ""
	for _, history := range changelog.Histories {
		_, err := db.Exec("INSERT INTO changelog_history VALUES ($1, $2, $3);",
			history.ID,
			issueKey,
			history.Created,
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		_, err = db.Exec("INSERT INTO changelog_history_author VALUES ($1, $2, $3, $4, $5, $6);",
			history.ID,
			history.Author.Name,
			history.Author.Email,
			history.Author.DisplayName,
			history.Author.Active,
			history.Author.TimeZone,
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		for _, historyItem := range history.Items {
			_, err := db.Exec("INSERT INTO changelog_history_item VALUES ($1, $2, $3, $4, $5, $6, $7);",
				history.ID,
				historyItem.Field,
				historyItem.FieldType,
				historyItem.From,
				historyItem.FromString,
				historyItem.To,
				historyItem.ToString,
			)
			if err != nil {
				errs += fmt.Sprintf("%s\n", err.Error())
			}
		}
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}
