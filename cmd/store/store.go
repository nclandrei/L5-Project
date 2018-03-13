package main

import (
	"flag"

	"github.com/nclandrei/L5-Project/db"

	"log"
	"math"
	"net/url"

	"github.com/nclandrei/L5-Project/jira"
)

// This defines the maximum number of concurrent client calls to Jira REST API
// as, otherwise, it would start dropping the connections
const maxNoGoroutines = 100

// store all the flags
var (
	jiraURLStr      = flag.String("jiraURL", "http://issues.apache.org", "the URL to the Jira instance")
	projectName     = flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	goroutinesCount = flag.Int("goroutinesCount", maxNoGoroutines, "defines the number of goroutines to be used")
	boltDBPath      = flag.String("dbPath", "users.db", "absolute path to the Bolt database")
)

func main() {
	flag.Parse()

	if *goroutinesCount > maxNoGoroutines {
		log.Fatalf("cannot have more than maximum number of goroutines... exiting now\n")
	}

	clientURL, err := url.Parse(*jiraURLStr)
	if err != nil {
		log.Fatalf("jira URL provided is not a valid URL: %v\n", err)
	}

	boltDB, err := db.NewBoltDB(*boltDBPath)
	if err != nil {
		log.Fatalf("could not create Bolt DB: %v\n", err)
	}

	// langClient, err := gcp.NewLanguageClient(context.Background())
	// if err != nil {
	// 	log.Fatalf("could not create GCP language client: %v\n", err)
	// }

	jiraClient, err := jira.NewClient(clientURL)
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

	done := make(chan *jira.SearchResponse)
	errs := make(chan error)
	dbChan := make(chan []jira.Issue)
	dbErrChan := make(chan error)

	var issues []jira.Issue

	for i := 0; i < *goroutinesCount; i++ {
		go jiraClient.GetPaginatedIssues(done, errs, i, int(issueSliceSize), *projectName)
	}

	// start goroutine listening for issues sent to the channel
	go boltDB.InsertIssues(dbChan, dbErrChan)

	go func() {
		for err := range dbErrChan {
			if err != nil {
				log.Println(err)
			}
		}
	}()

	for i := 0; i < 2*(*goroutinesCount); i++ {
		select {
		case searchResponse := <-done:
			if searchResponse != nil {
				dbChan <- searchResponse.Issues
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

	log.Printf("retrieved %d issues from Jira\n", len(issues))

	// close both database channels
	close(dbChan)
}
