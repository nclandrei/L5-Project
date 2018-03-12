package main

import (
	"flag"

	"github.com/nclandrei/L5-Project/db"

	"github.com/nclandrei/L5-Project/jira"
	// "github.com/nclandrei/L5-Project/processing"
	"log"
	"math"
	"net/url"
)

// This defines the maximum number of concurrent client calls to Jira REST API
// as, otherwise, it would start dropping the connections
const maxNoGoroutines = 100

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	goroutinesCount := flag.Int("goroutinesCount", 100, "defines the number of goroutines to be used")

	flag.Parse()

	if *goroutinesCount > maxNoGoroutines {
		log.Fatalf("cannot have more than maximum number of goroutines... exiting now")
	}

	jiraClient, err := jira.NewClient(&url.URL{
		Scheme: "http",
		Host:   "issues.apache.org",
	})
	if err != nil {
		log.Fatalf("Could not create Jira client: %v\n", err)
	}

	err = jiraClient.AuthenticateClient()
	if err != nil {
		log.Fatalf("Could not authenticate Jira client with Apache: %v\n", err)
	}

	numberOfIssues, err := jiraClient.GetNumberOfIssues(*projectName)
	if err != nil {
		log.Fatalf("Could not get total number of issues: %v\n", err)
	}

	issueSliceSize := math.Ceil(float64(numberOfIssues) / float64(*goroutinesCount))

	done := make(chan *jira.SearchResponse, numberOfIssues)
	errs := make(chan error, numberOfIssues)

	var issues []jira.Issue

	for i := 0; i < *goroutinesCount; i++ {
		go jiraClient.GetPaginatedIssues(done, errs, i, int(issueSliceSize), *projectName)
	}

	for i := 0; i < 2*(*goroutinesCount); i++ {
		select {
		case searchResponse := <-done:
			if searchResponse != nil {
				for _, issue := range searchResponse.Issues {
					issues = append(issues, issue)
				}
			}
		case err := <-errs:
			if err != nil {
				log.Printf("could not retrieve issues: %v\n", err)
			}
		}
	}

	log.Printf("got %d issues\n", len(issues))

	boltDB, err := db.NewBoltDB("users.db")
	if err != nil {
		log.Fatalf("could not create Bolt DB: %v\n", err)
	}

	err = boltDB.InsertIssues(issues)
	if err != nil {
		log.Fatalf("could not insert issues: %v\n", err)
	}

	dbIssues, err := boltDB.GetIssues()
	if err != nil {
		log.Fatalf("could not retrieve issues: %v\n", err)
	}

	log.Println(dbIssues)
}
