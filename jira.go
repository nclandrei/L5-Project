package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// SearchRequest defines what goes inside a JSON body for Jira JQL REST endpoint
type SearchRequest struct {
	Jql        string   `json:"jql,omitempty"`
	StartAt    int      `json:"startAt,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Fields     []string `json:"fields,omitempty"`
}

// SearchResponse defines the payload retrieved by Jira when a search is
// conducted via its REST API
type SearchResponse struct {
	Summary      string    `json:"summary"`
	Description  string    `json:"description"`
	TimeEstimate int       `json:"timeestimate"`
	TimeSpent    int       `json:"timespent"`
	Status       Status    `json:"status"`
	DueDate      time.Time `json:"duedate"`
	Progress     string    `json:"progress"`
	Comment      []Comment `json:"comment"`
	Priority     Priority  `json:"priority"`
	Key          string    `json:"key"`
	IssueType    IssueType `json:"issuetype"`
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

// NewSearchRequest returns a new initialized request
func NewSearchRequest(projectName string, paginationIndex, pageCount int) *SearchRequest {
	return &SearchRequest{
		Jql:        fmt.Sprintf("project=%s", projectName),
		StartAt:    paginationIndex * pageCount,
		MaxResults: pageCount,
		Fields: []string{"summary", "description", "comments", "key", "issuetype", "timespent",
			"priority", "timeestimate", "status", "duedate", "progress"},
	}
}

func getIssues(
	responses chan<- []byte,
	done chan<- bool,
	paginationIndex int,
	pageCount int,
	projectName string) {

	requestBody := NewSearchRequest(projectName, paginationIndex, pageCount)

	req, _ := json.Marshal(requestBody)

	resp, err := http.Post(jiraURL, "application/json", bytes.NewBuffer(req))
	if err != nil {
		log.Printf("Could not send request: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			if bodyJSON, err := json.Marshal(bodyBytes); err != nil {
				log.Printf("Could not marshal response to JSON: %v\n", err)
			} else {
				log.Println(bodyJSON)
				responses <- bodyJSON
			}
		}
	}
	done <- true
}
