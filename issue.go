package main

import (
	"time"
)

// Issue defines the Jira issue retrieved via the REST API
type Issue struct {
	Fields Fields `json:"fields,omitempty"`
}

// Fields defines the fields retrieved via the REST API
type Fields struct {
	Summary      string    `json:"summary,omitempty"`
	Description  string    `json:"description,omitempty"`
	TimeEstimate int       `json:"timeestimate,omitempty"`
	TimeSpent    int       `json:"timespent,omitempty"`
	Status       Status    `json:"status,omitempty"`
	DueDate      time.Time `json:"duedate,omitempty"`
	Comment      []Comment `json:"comment,omitempty"`
	Priority     Priority  `json:"priority,omitempty"`
	Key          string    `json:"key,omitempty"`
	IssueType    IssueType `json:"issuetype,omitempty"`
}

// IssueType defines the issue type in Jira
type IssueType struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Priority holds the type of priority assigned to a Jira issue
type Priority struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Status defines the Jira issue status
type Status struct {
	Name           string         `json:"name,omitempty"`
	StatusCategory StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory defines the category a Status belongs to (e.g. in progress)
type StatusCategory struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Comment defines the structure of a Jira issue comment
type Comment struct {
	Body    string        `json:"body,omitempty"`
	Author  CommentAuthor `json:"author"`
	Created time.Time     `json:"created,omitempty"`
	Updated time.Time     `json:"updated,omitempty"`
}

// CommentAuthor holds the name of a comment's author
type CommentAuthor struct {
	Name string `json:"name,omitempty"`
}
