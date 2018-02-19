package main

import (
	"flag"
	"github.com/nclandrei/L5-Project/db"
	"github.com/nclandrei/L5-Project/jira"
	"log"
)

// This defines the maximum number of concurrent client calls to Jira REST API
// as, otherwise, it would start dropping the connections
const maxNoGoroutines = 100

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	numberOfIssues := flag.Int("issuesCount", 1000000, "defines the number of issues to be retrieved")
	goroutinesCount := flag.Int("goroutinesCount", 100, "defines the number of goroutines to be used")

	flag.Parse()

	if *goroutinesCount > maxNoGoroutines {
		log.Fatalf("cannot have more than maximum number of goroutines... exitting now")
	}

	issuesPerPage := float64(*numberOfIssues) / float64(*goroutinesCount)

	done := make(chan *jira.SearchResponse, *numberOfIssues)
	errs := make(chan error, *numberOfIssues)
	var issues []jira.Issue

	jiraClient, err := jira.NewClient()
	if err != nil {
		log.Fatalf("Could not create Jira client: %v\n", err)
	}

	err = jiraClient.AuthenticateClient()

	if err != nil {
		log.Fatalf("Could not authenticate Jira client with Apache: %v\n", err)
	}

	for i := 0; i < *goroutinesCount; i++ {
		go jiraClient.GetPaginatedIssues(done, errs, i, int(issuesPerPage), *projectName)
	}

	for i := 0; i < *goroutinesCount; i++ {
		if searchResponse := <-done; searchResponse != nil {
			for _, issue := range searchResponse.Issues {
				issues = append(issues, issue)
			}
		}
		if err := <-errs; err != nil {
			log.Printf("Error while retrieving paginated issues: %v\n", err)
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
