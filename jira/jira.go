package jira

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

// Client defines the client for Jira
type Client struct {
	URL *url.URL
	*http.Client
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
	Expand     string  `json:"expand,omitempty"`
	StartAt    int     `json:"startAt,omitempty"`
	MaxResults int     `json:"maxResults,omitempty"`
	Total      int     `json:"total,omitempty"`
	Issues     []Issue `json:"issues,omitempty"`
}

// JiraSession represents a JiraSession JSON response by the JIRA API.
type JiraSession struct {
	Self    string `json:"self,omitempty"`
	Name    string `json:"name,omitempty"`
	Session struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"session,omitempty"`
	LoginInfo struct {
		FailedLoginCount    int    `json:"failedLoginCount"`
		LoginCount          int    `json:"loginCount"`
		LastFailedLoginTime string `json:"lastFailedLoginTime"`
		PreviousLoginTime   string `json:"previousLoginTime"`
	} `json:"loginInfo"`
	Cookies []*http.Cookie
}

// NewSearchRequest returns a new initialized request
func NewSearchRequest(projectName string, paginationIndex, pageCount int) *SearchRequest {
	return &SearchRequest{
		Jql:        fmt.Sprintf("project = %s", projectName),
		StartAt:    paginationIndex * pageCount,
		MaxResults: pageCount,
		Fields: []string{"summary", "description", "comments", "key", "issuetype", "timespent",
			"priority", "timeestimate", "status", "duedate", "progress"},
	}
}

// NewClient returns a new Jira Client
func NewClient() (*Client, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: &http.Client{
			Timeout: time.Second * 10,
			Jar:     cookieJar,
		},
		URL: &url.URL{
			Scheme: "https",
			Host:   "issues.apache.org",
		},
	}, nil
}

// AuthenticateClient authenticates a Jira client with a specific instance of Jira
func (client *Client) AuthenticateClient() error {
	authenticationRequest := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		os.Getenv("APACHE_JIRA_USERNAME"),
		os.Getenv("APACHE_JIRA_PASSWORD"),
	}

	client.URL.Path = "jira/rest/auth/1/session"

	jsonPayload, err := json.Marshal(authenticationRequest)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", client.URL.String(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	client.Jar.SetCookies(client.URL, response.Cookies())

	return nil
}

// GetPaginatedIssues adds to channels responses retrieved from Jira
func (client *Client) GetPaginatedIssues(
	responses chan<- *SearchResponse,
	errs chan<- error,
	paginationIndex int,
	pageCount int,
	projectName string) {

	searchRequestBody := NewSearchRequest(projectName, paginationIndex, pageCount)
	reqBody, err := json.Marshal(searchRequestBody)

	if err != nil {
		responses <- nil
		errs <- err
	}

	client.URL.Path = "jira/rest/api/2/search"

	request, err := http.NewRequest("POST", client.URL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		responses <- nil
		errs <- err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	resp, err := client.Do(request)

	if err != nil {
		responses <- nil
		errs <- err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			var searchResponse SearchResponse
			if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
				errs <- err
				responses <- nil
			} else {
				responses <- &searchResponse
				errs <- nil
			}
		} else {
			errs <- errors.New("Status code different than 200")
			responses <- nil
		}
	}
}
