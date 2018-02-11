package main

import (
	"time"
)

// Issue defines the Jira issue retrieved via the REST API
type Issue struct {
	Fields Fields `json:"fields"`
}

// Fields defines the fields retrieved via the REST API
type Fields struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	// TimeEstimate int       `json:"timeestimate"`
	// TimeSpent    int       `json:"timespent"`
	// Status       Status    `json:"status"`
	// DueDate      time.Time `json:"duedate"`
	// Progress     string    `json:"progress"`
	// Comment      []Comment `json:"comment"`
	// Priority     Priority  `json:"priority"`
	// Key          string    `json:"key"`
	// IssueType    IssueType `json:"issuetype"`
}

// IssueType defines the issue type in Jira
type IssueType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Priority holds the type of priority assigned to a Jira issue
type Priority struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Status defines the Jira issue status
type Status struct {
	Name string `json:"name"`
}

// Comment defines the structure of a Jira issue comment
type Comment struct {
	Body    string        `json:"body"`
	Author  CommentAuthor `json:"author"`
	Created time.Time     `json:"created"`
	Updated time.Time     `json:"updated"`
}

// CommentAuthor holds the name of a comment's author
type CommentAuthor struct {
	Name string `json:"name"`
}
