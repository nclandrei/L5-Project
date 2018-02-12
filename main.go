package main

import (
	"flag"
	"log"
)

// This defines the maximum number of concurrent client calls to Jira REST API
// as, otherwise, it would start dropping the connections
var maxNoGoroutines = 10

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	numberOfIssues := flag.Int("issuesCount", 2000, "defines the number of issues to be retrieved")
	goroutinesCount := flag.Int("goroutinesCount", 5, "defines the number of goroutines to be used")

	flag.Parse()

	if *goroutinesCount > maxNoGoroutines {
		log.Fatalf("cannot have more than maximum number of goroutines... exitting now")
	}

	issuesPerPage := float64(*numberOfIssues) / float64(*goroutinesCount)

	responses := make(chan SearchResponse)
	done := make(chan bool)
	var respSlice []SearchResponse

	jiraClient, err := NewJiraClient()
	if err != nil {
		log.Fatalf("Could not create Jira client: %v\n", err)
	}

	err = jiraClient.AuthenticateClient()

	if err != nil {
		log.Fatalf("Could not authenticate Jira client with Apache: %v\n", err)
	}

	for i := 0; i < *goroutinesCount; i++ {
		go jiraClient.GetPaginatedIssues(responses, done, i, int(issuesPerPage), *projectName)
	}

	doneCounter := 0

	for doneCounter < *goroutinesCount {
		select {
		case newResponse := <-responses:
			respSlice = append(respSlice, newResponse)
		case <-done:
			doneCounter++
		}
	}

	for _, value := range respSlice {
		for _, issue := range value.Issues {
			log.Printf("Key: " + issue.Key + "; Summary: " + issue.Fields.Summary + "\n")
		}
	}
}
