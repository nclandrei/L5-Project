package jira

import (
	"fmt"
	"strings"
	"time"
)

// Custom time format corresponding to Jira format
const timeFormat = "2006-01-02T15:04:05.000-0700"

// JTime holds the time formatted in Jira's specific format
type JTime time.Time

// AttachmentType defines the attachment file type inside Jira issues
type AttachmentType int

// UnmarshalJSON represents the formatting of JSON time for Jira's specific format
func (t *JTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "null" {
		*t = JTime(time.Time{})
		return nil
	}

	jiraTime, err := time.Parse(timeFormat, s)
	if err != nil {
		jiraTime, err = time.Parse("2006-01-02", s)
		if err != nil {
			return fmt.Errorf("could not parse JTime: %v", err)
		}
	}

	*t = JTime(jiraTime)

	return nil
}

// MarshalJSON marshals a JiraTime struct into a slice of bytes
func (t JTime) MarshalJSON() ([]byte, error) {
	jTime := fmt.Sprintf("\"%s\"", time.Time(t).Format(timeFormat))
	return []byte(jTime), nil
}

const (
	// Image type for attachments (e.g. png, jpg, jpeg)
	Image AttachmentType = iota + 1
	// Text type for attachments (e.g. txt, md)
	Text
	// Code type for attachments (e.g. go, java, clj)
	Code
	// Video type for attachments (e.g. mp4, mkv, avi)
	Video
)

// Issue defines the Jira issue retrieved via the REST API
type Issue struct {
	Key             string    `json:"key" bson:"_id"`
	Expand          string    `json:"_"`
	ID              string    `json:"-"`
	Self            string    `json:"-"`
	Fields          Fields    `json:"fields"`
	Changelog       Changelog `json:"changelog"`
	TimeToClose     float64
	SentimentScore  float32
	GrammarErrCount int
}

// Fields defines the fields retrieved via the REST API
type Fields struct {
	Summary      string       `json:"summary"`
	Description  string       `json:"description,omitempty"`
	TimeEstimate int          `json:"timeestimate,omitempty"`
	TimeSpent    int          `json:"timespent,omitempty"`
	Created      JTime        `json:"created"`
	Attachments  []Attachment `json:"attachment,omitempty"`
	Status       Status       `json:"status,omitempty"`
	DueDate      JTime        `json:"duedate,omitempty"`
	Comments     Comments     `json:"comment,omitempty"`
	Priority     Priority     `json:"priority,omitempty"`
	IssueType    IssueType    `json:"issuetype,omitempty"`
}

// Changelog defines the entire changelog of a Jira issue
type Changelog struct {
	StartAt    int                `json:"startAt"`
	MaxResults int                `json:"maxResults"`
	Total      int                `json:"total"`
	Histories  []ChangelogHistory `json:"histories,omitempty"`
}

// ChangelogHistory defines the entire history for some specific items
type ChangelogHistory struct {
	ID      string                 `json:"id,omitempty"`
	Author  Author                 `json:"author,omitempty"`
	Created JTime                  `json:"created,omitempty"`
	Items   []ChangelogHistoryItem `json:"items,omitempty"`
}

// Author holds the author for any jira issue field
type Author struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"emailAddress,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active,omitempty"`
	TimeZone    string `json:"timeZone,omitempty"`
}

// ChangelogHistoryItem defines a specific item inside a changelog history for a Jira issue
type ChangelogHistoryItem struct {
	Field      string `json:"field,omitempty"`
	FieldType  string `json:"fieldtype,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// Attachment defines a Jira attachment
type Attachment struct {
	ID       string `json:"id,omitempty"`
	Author   Author `json:"author,omitempty"`
	Filename string `json:"filename,omitempty"`
	Created  JTime  `json:"created,omitempty"`
	Size     int    `json:"size,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Content  string `json:"content,omitempty"`
}

// IssueType defines the issue type in Jira
type IssueType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Priority holds the type of priority assigned to a Jira issue
type Priority struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Status defines the Jira issue status
type Status struct {
	ID             string   `json:"id,omitempty"`
	Description    string   `json:"description,omitempty"`
	Name           string   `json:"name,omitempty"`
	StatusCategory struct{} `json:"-"`
}

// Comments defines the Jira field that holds the comments
type Comments struct {
	Comments []Comment `json:"comments,omitempty"`
}

// Comment defines the structure of a Jira issue comment
type Comment struct {
	ID      string `json:"id,omitempty"`
	Body    string `json:"body,omitempty"`
	Author  Author `json:"author"`
	Created JTime  `json:"created,omitempty"`
	Updated JTime  `json:"updated,omitempty"`
}
