package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
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
	numberOfIssues := flag.Int("issuesCount", 50000, "defines the number of issues to be retrieved")

	flag.Parse()

	responses := make(chan []byte)
	done := make(chan bool)
	var respSlice [][]byte

	for i := 0; i < *numberOfIssues/100; i++ {
		go getIssues(responses, done, i, 500, *projectName)
	}

	doneCounter := 0

	for doneCounter < *numberOfIssues/100 {
		select {
		case newResponse := <-responses:
			respSlice = append(respSlice, newResponse)
		case <-done:
			doneCounter++
		}
	}

	response := respSlice[0]
	var dat map[string]interface{}
	err := json.Unmarshal(response, &dat)
	if err != nil {
		log.Fatalf("Cannot parse JSON: %v", err)
	} else {
		fmt.Println(dat)
	}
}
