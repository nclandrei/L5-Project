package main

import (
	"flag"
	"github.com/nclandrei/L5-Project/db"
	"github.com/nclandrei/L5-Project/jira"
	"log"
)

// This defines the maximum number of concurrent client calls to Jira REST API
// as, otherwise, it would start dropping the connections
var maxNoGoroutines = 10

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	numberOfIssues := flag.Int("issuesCount", 1000, "defines the number of issues to be retrieved")
	goroutinesCount := flag.Int("goroutinesCount", 10, "defines the number of goroutines to be used")

	flag.Parse()

	if *goroutinesCount > maxNoGoroutines {
		log.Fatalf("cannot have more than maximum number of goroutines... exitting now")
	}

	issuesPerPage := float64(*numberOfIssues) / float64(*goroutinesCount)

	responses := make(chan jira.SearchResponse)
	done := make(chan bool)
	var issues []jira.Issue

	jiraClient, err := jira.NewJiraClient()
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
			for _, issue := range newResponse.Issues {
				issues = append(issues, issue)
			}
		case <-done:
			doneCounter++
		}
	}

	database, err := db.NewJiraDatabase()
	if err != nil {
		log.Fatalf("Could not create database: %v", err)
	}
	err = database.AddIssues(issues)
	if err != nil {
		log.Fatalf("Could not add issue to database: %v", err)
	}
}
