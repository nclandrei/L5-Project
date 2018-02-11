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

	for _, value := range respSlice {
		var dat []byte
		err := json.Unmarshal(value, dat)
		if err != nil {
			log.Fatalf("Cannot parse JSON: %v", err)
		} else {
			fmt.Println(string(dat))
		}
	}
}
