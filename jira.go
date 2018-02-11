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

// JiraClient defines the client for Jira
type JiraClient struct {
	URL        string
	HTTPClient *http.Client
}

// NewJiraClient returns a new Jira Client
func NewJiraClient() *JiraClient {
	return &JiraClient{
		URL: "http://issues.apache.org/jira/rest/api/2/search",
		HTTPClient: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

// SearchRequest defines what goes inside a JSON body for Jira JQL REST endpoint
type SearchRequest struct {
	Jql        string   `json:"jql,omitempty"`
	StartAt    int      `json:"startAt,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Fields     []string `json:"fields,omitempty"`
}

// SearchResponse defines the response payload retrieved through the search endpoint
type SearchResponse struct {
	Expand     string  `json:"expand"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// NewSearchRequest returns a new initialized request
func NewSearchRequest(projectName string, paginationIndex, pageCount int) *SearchRequest {
	return &SearchRequest{
		Jql:        fmt.Sprintf("project = %s", projectName),
		StartAt:    paginationIndex * pageCount,
		MaxResults: pageCount,
		Fields:     []string{"summary", "description"},
		// Fields: []string{"summary", "description", "comments", "key", "issuetype", "timespent",
		// 	"priority", "timeestimate", "status", "duedate", "progress"},
	}
}

// GetPaginatedIssues adds to channels responses retrieved from Jira
func (client *JiraClient) GetPaginatedIssues(
	responses chan<- Fields,
	done chan<- bool,
	paginationIndex int,
	pageCount int,
	projectName string) {

	searchRequestBody := NewSearchRequest(projectName, paginationIndex, pageCount)
	reqBody, err := json.Marshal(searchRequestBody)

	if err != nil {
		log.Fatalf("Could not encode search request to JSON: %v\n", err)
	}

	request, err := http.NewRequest("POST", client.URL, bytes.NewBuffer(reqBody))

	if err != nil {
		log.Fatalf("Could not create request: %v\n", err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	resp, err := client.HTTPClient.Do(request)

	if err != nil {
		log.Printf("Could not send request: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			var searchResponse Issue
			if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
				log.Printf("Could not marshal response to JSON: %v\n", err)
			} else {
				fmt.Println(searchResponse)
				responses <- searchResponse.Fields
			}
		}
	}
	done <- true
}
