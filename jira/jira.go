package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Client defines the client for Jira
type Client struct {
	URL *url.URL
	*http.Client
}

// SearchResponse defines the response payload retrieved through the search endpoint
type SearchResponse struct {
	Expand     string  `json:"expand,omitempty"`
	StartAt    int     `json:"startAt,omitempty"`
	MaxResults int     `json:"maxResults,omitempty"`
	Total      int     `json:"total,omitempty"`
	Issues     []Issue `json:"issues,omitempty"`
}

// Session represents a Session JSON response by the JIRA API.
type Session struct {
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

// NewClient returns a new Jira Client
func NewClient(url *url.URL) (*Client, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: &http.Client{
			Timeout: time.Second * 90,
			Jar:     cookieJar,
		},
		URL: url,
	}, nil
}

// setSearchPath sets the URL path for JQL search on a Jira client
func (client *Client) setSearchPath(projectName string, paginationIndex, pageCount int) {
	client.URL.Path = "/jira/rest/api/2/search"
	queryValues := make(url.Values)
	queryValues.Add("jql", fmt.Sprintf("project=%s", projectName))
	queryValues.Add("startAt", strconv.Itoa(paginationIndex*pageCount))
	queryValues.Add("maxResults", strconv.Itoa(pageCount))
	queryValues.Add("fields", "summary, description, comments, key, issuetype, timespent, priority, timeestimate, status, duedate, progress")
	queryValues.Add("expand", "changelog")
	client.URL.RawQuery = queryValues.Encode()
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

	client.URL.Path = "/jira/rest/auth/1/session"

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

	client.setSearchPath(projectName, paginationIndex, pageCount)
	resp, err := client.Get(client.URL.String())

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
				errs <- nil
				responses <- &searchResponse
			}
		} else {
			errs <- fmt.Errorf("Status code different than 200: %v", resp.Status)
			responses <- nil
		}
	}
}
