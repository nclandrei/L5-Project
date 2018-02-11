package main

import (
	"time"
)

// Issue defines the Jira issue retrieved via the REST API
type Issue struct {
	Key    string `json:"key,omitempty"`
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
	IssueType    IssueType `json:"issuetype,omitempty"`
}

// IssueType defines the issue type in Jira
type IssueType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"self,omitempty"`
	Description string `json:"description,omitempty"`
	SubTask     bool   `json:"subtask,omitempty"`
	AvatarID    int    `json:"avatarId,omitempty"`
}

// Priority holds the type of priority assigned to a Jira issue
type Priority struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	IconURL string `json:"iconurl,omitempty"`
	URL     string `json:"self,omitempty"`
}

// Status defines the Jira issue status
type Status struct {
	URL            string         `json:"self,omitempty"`
	Description    string         `json:"description,omitempty"`
	IconURL        string         `json:"iconurl,omitempty"`
	ID             string         `json:"id,omitempty"`
	Name           string         `json:"name,omitempty"`
	StatusCategory StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory defines the category a Status belongs to (e.g. in progress)
type StatusCategory struct {
	URL       string `json:"self,omitempty"`
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Key       string `json:"key,omitempty"`
	ColorName string `json:"colorName,omitempty"`
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
