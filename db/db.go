package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // driver required for comminucating with postgres DB
	"github.com/nclandrei/L5-Project/jira"
	"time"
)

const (
	connection   = "user=nclandrei password=nclandrei dbname=nclandrei sslmode=disable"
	databaseType = "postgres"
	pgTimeFormat = "2006-01-02 15:04:05 MST"
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
	rows, err := db.Query("SELECT * FROM ISSUE;")
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
			&issue.Fields.DueDate,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, *issue)
	}
	return issues, nil
}

// InsertIssues inserts a slice of issues into the issues table
func (db *JiraDatabase) InsertIssues(project string, issues []jira.Issue) error {
	var errs string
	for _, issue := range issues {
		errs := ""
		_, err := db.Exec("INSERT INTO issue VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);",
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.TimeSpent,
			issue.Fields.TimeEstimate,
			time.Time(issue.Fields.DueDate).UTC().Format(pgTimeFormat),
			project,
			time.Time(issue.Fields.Created).UTC().Format(pgTimeFormat),
			jira.CalculateNumberOfWords(issue.Fields.Description),
			jira.CalculateNumberOfWords(issue.Fields.Summary),
		)
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
		err = insertComments(db, issue.Key, issue.Fields.Comments.Comments)
		if err != nil {
			errs += fmt.Sprintf("Could not insert comments for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertAttachments(db, issue.Key, issue.Fields.Attachments)
		if err != nil {
			errs += fmt.Sprintf("Could not insert attachments for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertStatus(db, issue.Key, issue.Fields.Status)
		if err != nil {
			errs += fmt.Sprintf("Could not insert status for issue %s: %s\n", issue.Key, err.Error())
		}
		err = insertChangelog(db, time.Time(issue.Fields.Created), issue.Key, issue.Changelog)
		if err != nil {
			errs += fmt.Sprintf("Could not insert changelog for issue %s: %s\n", issue.Key, err.Error())
		}
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}

func insertComments(db *JiraDatabase, issueKey string, comments []jira.Comment) error {
	errs := ""
	for _, comment := range comments {
		_, err := db.Exec("INSERT INTO comment VALUES ($1, $2, $3, $4, $5, $6);",
			comment.ID,
			issueKey,
			comment.Body,
			time.Time(comment.Created).UTC().Format(pgTimeFormat),
			time.Time(comment.Updated).UTC().Format(pgTimeFormat),
			jira.CalculateNumberOfWords(comment.Body),
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		_, err = db.Exec("INSERT INTO comment_author VALUES ($1, $2, $3);",
			comment.ID,
			issueKey,
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

func insertAttachments(db *JiraDatabase, issueKey string, attachments []jira.Attachment) error {
	errs := ""
	for _, attachment := range attachments {
		_, err := db.Exec("INSERT INTO attachment VALUES ($1, $2, $3, $4, $5, $6, $7, $8);",
			attachment.ID,
			issueKey,
			attachment.Filename,
			time.Time(attachment.Created).UTC().Format(pgTimeFormat),
			attachment.Size,
			attachment.MimeType,
			attachment.Content,
			int(jira.GetAttachmentType(attachment.Filename)),
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		_, err = db.Exec("INSERT INTO attachment_author VALUES ($1, $2, $3, $4, $5, $6, $7);",
			attachment.ID,
			issueKey,
			attachment.Author.Name,
			attachment.Author.Email,
			attachment.Author.DisplayName,
			attachment.Author.Active,
			attachment.Author.TimeZone,
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
	_, err := db.Exec("INSERT INTO status VALUES ($1, $2, $3, $4);",
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

func insertChangelog(db *JiraDatabase, issueCreatedTS time.Time, issueKey string, changelog jira.Changelog) error {
	errs := ""
	for _, history := range changelog.Histories {
		changelogCreatedTS := time.Time(history.Created).UTC()
		_, err := db.Exec("INSERT INTO changelog_history VALUES ($1, $2, $3, $4);",
			history.ID,
			issueKey,
			changelogCreatedTS.Format(pgTimeFormat),
			jira.CalculateJTimeDifference(changelogCreatedTS, issueCreatedTS),
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		_, err = db.Exec("INSERT INTO changelog_history_author VALUES ($1, $2, $3, $4, $5, $6, $7);",
			history.ID,
			history.Author.Name,
			history.Author.Email,
			history.Author.DisplayName,
			history.Author.Active,
			history.Author.TimeZone,
			issueKey,
		)
		if err != nil {
			errs += fmt.Sprintf("%s\n", err.Error())
		}
		for _, historyItem := range history.Items {
			_, err := db.Exec("INSERT INTO changelog_history_item VALUES ($1, $2, $3, $4, $5, $6, $7, $8);",
				history.ID,
				historyItem.Field,
				historyItem.FieldType,
				historyItem.From,
				historyItem.FromString,
				historyItem.To,
				historyItem.ToString,
				issueKey,
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
