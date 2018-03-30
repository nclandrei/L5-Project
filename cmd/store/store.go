package main

import (
	"context"
	"flag"
	"sync"

	"github.com/nclandrei/L5-Project/analyze"

	"github.com/nclandrei/L5-Project/gcp"

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
	dbPath          = flag.String("dbPath", "users.db", "absolute path to the Bolt database")
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

	boltDB, err := db.NewBoltDB(*dbPath)
	if err != nil {
		log.Fatalf("could not create Bolt DB: %v\n", err)
	}

	langClient, err := gcp.NewLanguageClient(context.Background())
	if err != nil {
		log.Fatalf("could not create GCP language client: %v\n", err)
	}

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
	preprocessCh := make(chan []jira.Issue)
	dbCh := make(chan []jira.Issue)
	pipelineDone := make(chan struct{})

	go func() {
		for ii := range preprocessCh {
			for i := range ii {
				bi, err := boltDB.IssueByKey(ii[i].Key)
				if err != nil {
					log.Printf("could not retrieve issue {%s} from bolt: %v\n", ii[i].Key, err)
					continue
				}
				if bi != nil || bi.CommSentiment != 0 {
					continue
				}
				concatComm, err := analyze.ConcatenateComments(ii[i])
				if err != nil {
					log.Printf("could not concatenate comments for issue {%s}: %v\n", ii[i].Key, err)
					continue
				}
				score, err := langClient.CommSentimentScore(concatComm)
				if err != nil {
					log.Printf("could not calculate sentiment score for issue {%s}: %v\n", ii[i].Key, err)
					continue
				}
				ii[i].CommSentiment = score
			}
			dbCh <- ii
		}
	}()

	go func() {
		for issues := range dbCh {
			err := boltDB.InsertIssues(issues...)
			if err != nil {
				log.Printf("could not add issues to bolt: %v\n", err)
			}
		}
		pipelineDone <- struct{}{}
	}()

	var wg sync.WaitGroup

	for i := 0; i < *goroutinesCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			issueSlice, err := jiraClient.GetIssues(*projectName, index, int(issueSliceSize))
			if err != nil {
				log.Printf("error while getting issues: %v\n", err)
				return
			}
			preprocessCh <- issueSlice
		}(i)
	}

	wg.Wait()

	close(dbCh)
	close(preprocessCh)

	<-pipelineDone
}
