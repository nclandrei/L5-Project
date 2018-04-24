package main

import (
	"flag"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nclandrei/ticketguru/db"

	"log"
	"math"
	"net/url"

	"github.com/nclandrei/ticketguru/jira"
)

// This defines the maximum number of concurrent client calls to Jira REST API
// as, otherwise, it would start dropping the connections
const maxNoGoroutines = 100

var (
	jiraURL     = flag.String("jiraURL", "http://issues.apache.org", "URL for Jira instance")
	project     = flag.String("project", "Kafka", "name of the project to be queried upon")
	gortnCnt    = flag.Int("goroutinesCount", maxNoGoroutines, "number of goroutines to be used")
	dbPath      = flag.String("dbPath", "issues.db", "absolute path to the Bolt database")
	logToFile   = flag.Bool("file_log", false, "specifies whether application should log to file or not")
	logFilePath = flag.String("log_path", "~/Code/go/src/github.com/nclandrei/ticketguru/log.txt", "path to logging file")
)

func main() {
	flag.Parse()

	var logger *log.Logger

	if !*logToFile {
		logger = log.New(os.Stdout, "jira-store: ", log.Lshortfile)
	} else {
		_, err := os.Stat(*logFilePath)
		if os.IsNotExist(err) {
			file, err := os.Create(*logFilePath)
			if err != nil {
				log.Fatalf("could not create logging file: %v\n", err)
			}
			logger = log.New(file, "jira-store: ", log.Lshortfile)
		}
	}

	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interruptCh
		log.SetOutput(os.Stderr)
		log.Printf("interrupt issued... exitting now...")
		os.Exit(1)
	}()

	err := godotenv.Load()
	if err != nil {
		logger.Fatalf("could not load .env file: %v\n", err)
	}

	if *gortnCnt > maxNoGoroutines {
		logger.Fatalf("cannot have more than maximum number of goroutines... exiting now\n")
	}

	clientURL, err := url.Parse(*jiraURL)
	if err != nil {
		logger.Fatalf("jira URL provided is not a valid URL: %v\n", err)
	}

	jiraClient, err := jira.NewClient(clientURL)
	if err != nil {
		logger.Fatalf("could not create Jira client: %v\n", err)
	}

	boltDB, err := db.NewBolt(*dbPath)
	if err != nil {
		logger.Fatalf("could not create Bolt DB: %v\n", err)
	}

	err = jiraClient.AuthenticateClient()
	if err != nil {
		logger.Fatalf("could not authenticate Jira client: %v\n", err)
	}

	numberOfIssues, err := jiraClient.TicketsCount(*project)
	if err != nil {
		logger.Fatalf("could not get total number of tickets: %v\n", err)
	}

	issueSliceSize := math.Ceil(float64(numberOfIssues) / float64(*gortnCnt))

	var wg sync.WaitGroup

	for i := 0; i < *gortnCnt; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			issues, err := jiraClient.Tickets(*project, index, int(issueSliceSize))
			if err != nil {
				logger.Printf("error while getting issues: %v\n", err)
			}
			err = boltDB.Insert(issues...)
			if err != nil {
				logger.Printf("could not add issues to bolt: %v\n", err)
			}
		}(i)
	}

	wg.Wait()
}
