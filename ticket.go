package ticketguru

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// Custom time format corresponding to Jira format.
	timeFormat = "2006-01-02T15:04:05.000-0700"

	// MaxTimeToCloseH represents the maximum number of hours until ticket closing allowed in analysis, plotting and stats.
	MaxTimeToCloseH = 27000

	// MaxCommWordCount represents the maximum number of comments allowed in analysis, plotting and stats.
	MaxCommWordCount = 25000

	// MaxGrammarErrCount represents the maximum number of grammar errors allowed in analysis, plotting and stats.
	MaxGrammarErrCount = 115

	// MaxSummaryDescWordCount represents the maximum number of summary & description words allowed in
	// analysis, plotting and stats.
	MaxSummaryDescWordCount = 5000
)

// Time holds the time formatted in Jira's specific format.
type Time time.Time

// UnmarshalJSON represents the formatting of JSON time for Jira's specific format
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "null" {
		*t = Time(time.Time{})
		return nil
	}

	jiraTime, err := time.Parse(timeFormat, s)
	if err != nil {
		jiraTime, err = time.Parse("2006-01-02", s)
		if err != nil {
			return fmt.Errorf("could not parse JTime: %v", err)
		}
	}

	*t = Time(jiraTime)

	return nil
}

// MarshalJSON marshals a JiraTime struct into a slice of bytes
func (t Time) MarshalJSON() ([]byte, error) {
	jTime := fmt.Sprintf("\"%s\"", time.Time(t).Format(timeFormat))
	return []byte(jTime), nil
}

// JiraIssue defines a Jira ticket.
type JiraIssue struct {
	Key                   string    `json:"key" bson:"_id"`
	Expand                string    `json:"_"`
	ID                    string    `json:"-"`
	Self                  string    `json:"-"`
	Fields                Fields    `json:"fields"`
	Changelog             Changelog `json:"changelog"`
	TimeToClose           float64
	Sentiment             Sentiment
	GrammarCorrectness    GrammarCorrectness
	HasStackTrace         bool
	HasStepsToReproduce   bool
	SummaryDescWordsCount int
	CommentWordsCount     int
}

// Sentiment holds information regarding the sentiment analysis score and if the analysis has been conducted.
type Sentiment struct {
	Score    float64
	HasScore bool
}

// GrammarCorrectness holds information regarding the grammar correctness score and if the analysis has been conducted.
type GrammarCorrectness struct {
	Score    int
	HasScore bool
}

// Fields defines the fields retrieved via the REST API
type Fields struct {
	Summary      string       `json:"summary"`
	Description  string       `json:"description,omitempty"`
	TimeEstimate int          `json:"timeestimate,omitempty"`
	TimeSpent    int          `json:"timespent,omitempty"`
	Created      Time         `json:"created"`
	Attachments  []Attachment `json:"attachment,omitempty"`
	Status       Status       `json:"status,omitempty"`
	DueDate      Time         `json:"duedate,omitempty"`
	Comments     Comments     `json:"comment,omitempty"`
	Priority     Priority     `json:"priority,omitempty"`
	Type         Type         `json:"issuetype,omitempty"`
}

// TicketKey returns the unique key of a Jira issue.
func (t *JiraIssue) TicketKey() string {
	return t.Key
}

// TicketBody returns the JSON encoded value of a Jira issue.
func (t *JiraIssue) TicketBody() ([]byte, error) {
	res, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Changelog defines the entire changelog of a Jira ticket.
type Changelog struct {
	StartAt    int                `json:"startAt"`
	MaxResults int                `json:"maxResults"`
	Total      int                `json:"total"`
	Histories  []ChangelogHistory `json:"histories,omitempty"`
}

// ChangelogHistory defines the entire history for some specific items.
type ChangelogHistory struct {
	ID      string                 `json:"id,omitempty"`
	Author  Author                 `json:"author,omitempty"`
	Created Time                   `json:"created,omitempty"`
	Items   []ChangelogHistoryItem `json:"items,omitempty"`
}

// Author holds the author for any jira ticket field.
type Author struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"emailAddress,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active,omitempty"`
	TimeZone    string `json:"timeZone,omitempty"`
}

// ChangelogHistoryItem defines a specific item inside a changelog history for a Jira ticket.
type ChangelogHistoryItem struct {
	Field      string `json:"field,omitempty"`
	FieldType  string `json:"fieldtype,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// Attachment defines a Jira attachment.
type Attachment struct {
	ID       string         `json:"id,omitempty"`
	Author   Author         `json:"author,omitempty"`
	Filename string         `json:"filename,omitempty"`
	Created  Time           `json:"created,omitempty"`
	Size     int            `json:"size,omitempty"`
	MimeType string         `json:"mimeType,omitempty"`
	Content  string         `json:"content,omitempty"`
	Type     AttachmentType `json:"attachment_type,omitempty"`
}

// AttachmentType maps the extension of the attachment to a predefined type (e.g. image).
type AttachmentType int

const (
	// ImageAttachment represents the image type of an attachment (e.g. png, jpg).
	ImageAttachment AttachmentType = iota + 1
	// VideoAttachment represents the video type of an attachmnet (e.g. mkv, mp4, avi).
	VideoAttachment
	// CodeAttachment represents the code snippet type of an attachmnet (e.g. go, java, groovy).
	CodeAttachment
	// SpreadsheetAttachment represents the spreadsheet type of an attachmnet (e.g. numbers, csv, xlsx).
	SpreadsheetAttachment
	// TextAttachment represents the text type of an attachment (e.g. md, txt, org).
	TextAttachment
	// ConfigAttachment represents the configuration type of an attachment (e.g. json, xml, yaml).
	ConfigAttachment
	// ArchiveAttachment represents the archive type of an attachment (e.g. zip, tar, rar).
	ArchiveAttachment
	// OtherAttachment represents any other extension of the attachment that is not relevant to the analysis.
	OtherAttachment
)

// Type defines the type of a ticket in Jira.
type Type struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Priority holds the type of priority assigned to a Jira ticket.
type Priority struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Status defines the Jira ticket status.
type Status struct {
	ID             string   `json:"id,omitempty"`
	Description    string   `json:"description,omitempty"`
	Name           string   `json:"name,omitempty"`
	StatusCategory struct{} `json:"-"`
}

// Comments defines the Jira field that holds the comments.
type Comments struct {
	Comments []Comment `json:"comments,omitempty"`
}

// Comment defines the structure of a Jira ticket comment.
type Comment struct {
	ID      string `json:"id,omitempty"`
	Body    string `json:"body,omitempty"`
	Author  Author `json:"author"`
	Created Time   `json:"created,omitempty"`
	Updated Time   `json:"updated,omitempty"`
}

// IsHighPriority returns whether a ticket is of high priority or not.
func IsHighPriority(t JiraIssue) bool {
	if t.Fields.Priority.ID == "" {
		return false
	}
	pID, _ := strconv.Atoi(t.Fields.Priority.ID)
	return pID <= 4
}

// Ticket describes a general interface for either Jira issues or Bugzilla tickets.
type Ticket interface {
	TicketKey() string
	TicketBody() ([]byte, error)
}
