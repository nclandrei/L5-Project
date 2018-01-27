package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

const jiraURL = "https://issues.apache.org/jira/rest/api/2/search"

// JiraIssue represents an issue returned via Jira's REST API
type JiraIssue struct {
	summary  string
	comments []Comment
}

// Comment defines the structure of a Jira issue comment
type Comment struct {
	author string
	body   string
}

// JqlRequestBody defines what goes inside a JSON body for Jira JQL REST endpoint
type JqlRequestBody struct {
	Jql        string   `json:"jql,omitempty"`
	StartAt    int      `json:"startAt,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Expand     []string `json:"expand"`
}

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	numberOfIssues := flag.Int("issuesCount", 5000, "defines the number of issues to be retrieved")

	requestBody := &JqlRequestBody{
		Jql:        fmt.Sprintf("project=%v", projectName),
		StartAt:    0,
		MaxResults: *numberOfIssues,
		Expand:     []string{"summary", "comments"},
	}

	if req, err := json.Marshal(requestBody); err == nil {
		resp, err := http.Post(jiraURL, "application/json", bytes.NewBuffer(req))
		if err != nil {
			fmt.Printf("Could not send request: %v", err)
		} else {
			fmt.Println("response Status:", resp.Status)
		}
	} else {
		fmt.Printf("Could not marshal request body: %v", err)
	}
}
