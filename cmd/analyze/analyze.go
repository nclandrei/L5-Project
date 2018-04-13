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
		fmt.Printf("%s is not a valid analysis type; available types are langTool, bing, sentiment and all", analysisType)
		os.Exit(1)
	}

	cursor, teardown, err := boltDB.Cursor()
	if err != nil {
		log.Fatalf("could not retrieve issues from Bolt DB: %v\n", err)
	}

	var issues []jira.Issue

	_, v := cursor.First()

	for i := 0; i < 10; i++ {
		var issue jira.Issue
		err := json.Unmarshal(v, &issue)
		if err != nil {
			log.Fatalf("could not unmarshal issue from bolt db: %v\n", err)
		}
		issues = append(issues, issue)
		fmt.Println(issue.Key)
		_, v = cursor.Next()
	}

	err = teardown()
	if err != nil {
		log.Fatalf("could not close bolt DB tx: %v\n", err)
	}

	scoreMap, err := analyze.MultipleScores(issues, clients...)
	if err != nil {
		log.Fatalf("could not calculate scores: %v\n", err)
	}

	for k, v := range scoreMap {
		switch k {
		case "SENTIMENT":
			for i := range v {
				issues[i].SentimentScore = v[i]
			}
			break
		case "GRAMMAR":
			for i := range v {
				issues[i].GrammarErrCount = v[i]
			}
			break
		}
		err = boltDB.InsertIssues(issues...)
		if err != nil {
			log.Fatalf("could not insert issues back into db: %v\n", err)
		}
	}
}
