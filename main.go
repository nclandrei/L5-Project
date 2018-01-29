package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
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
	numberOfIssues := flag.Int("issuesCount", 50000, "defines the number of issues to be retrieved")

	flag.Parse()

	responses := make(chan []byte)
	done := make(chan bool)
	var respSlice [][]byte

	for i := 0; i < *numberOfIssues/100; i++ {
		go func(j int) {
			fmt.Println(j)
			requestBody := &JqlRequestBody{
				Jql:        fmt.Sprintf("project=%s", *projectName),
				StartAt:    j * 500,
				MaxResults: 500,
			}

			req, _ := json.Marshal(requestBody)

			resp, err := http.Post(jiraURL, "application/json", bytes.NewBuffer(req))
			if err != nil {
				fmt.Printf("Could not send request: %v", err)
			} else {
				fmt.Println("response Status:", resp.Status)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					bodyBytes, _ := ioutil.ReadAll(resp.Body)
					if bodyJSON, err := json.Marshal(bodyBytes); err != nil {
						fmt.Printf("Could not marshal response to JSON: %v", err)
					} else {
						responses <- bodyJSON
					}
				}
			}
			done <- true
		}(i)
	}

	doneCounter := 0

	for doneCounter < *numberOfIssues/100 {
		select {
		case newResponse := <-responses:
			respSlice = append(respSlice, newResponse)
		case <-done:
			doneCounter++
			fmt.Println("Worker finished executing")
		}
	}

	connStr := "user=nclandrei password=nclandrei dbname=nclandrei sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	rows, err := db.Query("SELECT * FROM ISSUES;")
	if err != nil {
		log.Fatalf("Could not query database for issues: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id sql.NullInt64
		var summary sql.NullString
		var description sql.NullString
		var comments sql.NullString
		var key string
		err = rows.Scan(&id, &summary, &description, &comments, &key)
		fmt.Printf("%v | %v | %v | %v | %v\n", id, summary, description, comments, key)
	}
}
