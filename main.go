package main

import (
	"flag"
	"github.com/nclandrei/L5-Project/db"
	"github.com/nclandrei/L5-Project/jira"
	"sync"
	// "github.com/nclandrei/L5-Project/plot"
	// "io/ioutil"
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

	issuesPerPage := math.Ceil(float64(numberOfIssues) / float64(*goroutinesCount))

	done := make(chan *jira.SearchResponse, numberOfIssues)
	errs := make(chan error, numberOfIssues)
	var issues []jira.Issue

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
			log.Printf("could not issues: %v\n", err)
		}
	}

	log.Printf("finished getting the issues from Jira; number of issues: %v\n", len(issues))

	// database, err := db.NewJiraDatabase()
	// if err != nil {
	// 	log.Fatalf("could not create database: %v", err)
	// }
	// err = database.InsertIssues(*projectName, issues)
	// if err != nil {
	// 	log.Fatalf("could not add issue to database: %v", err)
	// }

	// withAtchQuery, err := ioutil.ReadFile("resources/with_attachments.sql")
	// if err != nil {
	// 	log.Fatalf("could not read file: %v\n", err)
	// }

	// dbIssues, err := database.ExecuteQuery(string(withAtchQuery))
	// if err != nil {
	// 	log.Fatalf("could not retrieve issues from querying database: %v\n", err)
	// }

	// log.Printf("%v\n", dbIssues)

	// withoutAtchQuery, err := ioutil.ReadFile("resources/without_attachments.sql")
	// if err != nil {
	// 	log.Fatalf("could not read file: %v\n", err)
	// }

	// dbIssues, err = database.ExecuteQuery(string(withoutAtchQuery))
	// if err != nil {
	// 	log.Fatalf("could not retrieve issues from querying database: %v\n", err)
	// }

	// log.Printf("%v\n", dbIssues)

	db, err := db.NewDatabase("localhost", "nclandrei", "issues")
	if err != nil {
		log.Fatalf("could not retrieve mongo session: %v\n", err)
	}

	defer db.Session.Close()

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i jira.Issue) {
			defer wg.Done()
			err := db.InsertIssue(i)
			if err != nil {
				log.Printf("could not insert issue: %v\n", err)
			}
		}(issues[i])
	}

	wg.Wait()

	// plotter, err := plot.NewPlotter()
	// if err != nil {
	// 	log.Fatalf("could not create plotter: %s\n", err)
	// }
	// err = plotter.Draw("Attachments", "Time-To-Completion", "#Attachments", nil)
	// if err != nil {
	// 	log.Printf("could not draw points inside plotter: %v\n", err)
	// }
}
