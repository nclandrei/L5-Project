package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nclandrei/L5-Project/analyze"
	"github.com/nclandrei/L5-Project/db"
	"github.com/nclandrei/L5-Project/jira"
	"log"
	"os"
)

func main() {
	boltDB, err := db.NewBoltDB("users.db")
	if err != nil {
		log.Fatalf("could not access Bolt DB: %v\n", err)
	}

	var analysisType string
	flag.StringVar(&analysisType, "type", "all", "type of analysis to run; available types: grammar, sentiment, all")

	flag.Parse()

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env file: %v\n", err)
	}

	var clients []analyze.Scorer

	switch analysisType {
	case "grammar":
		clients = append(clients, analyze.NewBingClient(os.Getenv("BING_KEY_1")))
		break
	case "sentiment":
		sentimentClient, err := analyze.NewSentimentClient(context.Background())
		if err != nil {
			log.Fatalf("could not create GCP sentiment client: %v\n", err)
		}
		clients = append(clients, sentimentClient)
		break
	case "all":
		sentimentClient, err := analyze.NewSentimentClient(context.Background())
		if err != nil {
			log.Fatalf("could not create GCP sentiment client: %v\n", err)
		}
		clients = append(
			clients,
			sentimentClient,
			analyze.NewBingClient(os.Getenv("BING_KEY_1")),
		)
		break
	default:
		fmt.Printf("%s is not a valid analysis type; available types are grammar, sentiment and all", analysisType)
		os.Exit(1)
	}

	totalIssueLen, err := boltDB.IssueBucketSize()
	if err != nil {
		log.Fatalf("could not retrieve issues bucket size: %v\n", err)
	}

	cursor, teardown, err := boltDB.Cursor()
	if err != nil {
		log.Fatalf("could not retrieve bolt cursor: %v\n", err)
	}

	fmt.Println(totalIssueLen)
	os.Exit(1)
	sliceSize := 10000
	issues := make([]jira.Issue, sliceSize)

	_, v := cursor.First()
	for i := 0; i < totalIssueLen; i += totalIssueLen {
		for j := 0; j < sliceSize; j++ {
			var issue jira.Issue
			err := json.Unmarshal(v, &issue)
			if err != nil {
				log.Fatalf("could not json unmarshal issue: %v\n", err)
			}
			issues[j] = issue
			_, v = cursor.Next()
		}
		err = analyze.MultipleScores(issues, clients...)
		if err != nil {
			log.Printf("could not calculate scores: %v\n", err)
		}

		err = boltDB.InsertIssues(issues...)
		if err != nil {
			log.Fatalf("could not insert issues in db: %v\n", err)
		}
	}

	err = teardown()
	if err != nil {
		log.Fatalf("could not close bolt transaction: %v\n", err)
	}
}
