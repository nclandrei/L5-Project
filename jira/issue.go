package jira

// Issue defines the Jira issue retrieved via the REST API
type Issue struct {
	Key       string    `json:"key,omitempty"`
	Fields    Fields    `json:"fields,omitempty"`
	Changelog Changelog `json:"changelog,omitempty"`
}

// Fields defines the fields retrieved via the REST API
type Fields struct {
	Summary      string    `json:"summary,omitempty"`
	Description  string    `json:"description,omitempty"`
	TimeEstimate int       `json:"timeestimate,omitempty"`
	TimeSpent    int       `json:"timespent,omitempty"`
	Created      string    `json:"created,omitempty"`
	Status       Status    `json:"status,omitempty"`
	DueDate      string    `json:"duedate,omitempty"`
	Comments     Comments  `json:"comment,omitempty"`
	Priority     Priority  `json:"priority,omitempty"`
	IssueType    IssueType `json:"issuetype,omitempty"`
}

// Changelog defines the entire changelog of a Jira issue
type Changelog struct {
	StartAt    int                `json:"startAt,omitempty"`
	MaxResults int                `json:"maxResults,omitempty"`
	Total      int                `json:"total,omitempty"`
	Histories  []ChangelogHistory `json:"histories,omitempty"`
}

// ChangelogHistory defines the entire history for some specific items
type ChangelogHistory struct {
	ID      string                 `json:"id,omitempty"`
	Author  ChangelogHistoryAuthor `json:"author,omitempty"`
	Created string                 `json:"created,omitempty"`
	Items   []ChangelogHistoryItem `json:"items,omitempty"`
}

// ChangelogHistoryAuthor holds the author of a changelog history item
type ChangelogHistoryAuthor struct {
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
	ID      string        `json:"id,omitempty"`
	Body    string        `json:"body,omitempty"`
	Author  CommentAuthor `json:"author"`
	Created string        `json:"created,omitempty"`
	Updated string        `json:"updated,omitempty"`
}

// CommentAuthor holds the name of a comment's author
type CommentAuthor struct {
	Name string `json:"name,omitempty"`
}
