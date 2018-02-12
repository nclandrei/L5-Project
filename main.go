package main

import (
	"flag"
	"github.com/nclandrei/L5-Project/jira"
	"github.com/nclandrei/L5-Project/processing"
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
	var respSlice []jira.SearchResponse

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
			respSlice = append(respSlice, newResponse)
		case <-done:
			doneCounter++
		}
	}

	var counter int

	for _, value := range respSlice {
		counter += len(value.Issues)
		for _, issue := range value.Issues {
			log.Printf("Key: " + issue.Key + "; Summary: " + issue.Fields.Summary + "\n")
		}
	}

	if sentimentScore, err := processing.SentimentScoreFromDoc(respSlice[0].Issues[0].Fields.Summary); err != nil {
		log.Printf("Could not calculate sentiment score: %v\n", err)
	} else {
		log.Printf("Score is: %v\n", sentimentScore)
	}
}
