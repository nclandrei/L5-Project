package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nclandrei/L5-Project/analyze"
	"github.com/nclandrei/L5-Project/db"
	"log"
	"os"
)

func main() {
	boltDB, err := db.NewBoltDB("users.db")
	if err != nil {
		log.Fatalf("could not access Bolt DB: %v\n", err)
	}

	flag.Parse()

	var analysisType string
	clients := make([]analyze.Scorer, 2)
	flag.StringVar(&analysisType, "type", "all", "type of analysis to run; available types: grammar,"+
		" sentiment, all (sentiment and grammar)")

	switch analysisType {
	case "grammar":
		clients = append(clients, analyze.NewGrammarClient())
		break
	case "sentiment":
		sentimentClient, err := analyze.NewSentimentClient(context.Background())
		if err != nil {
			log.Fatalf("could not create GCP sentiment client: %v\n", err)
		}
		clients = append(clients, sentimentClient)
		break
	case "spellCheck":
		clients = append(clients, analyze.NewBingClient(os.Getenv("BING_KEY_1")))
		break
	case "all":
		sentimentClient, err := analyze.NewSentimentClient(context.Background())
		if err != nil {
			log.Fatalf("could not create GCP sentiment client: %v\n", err)
		}
		clients = append(clients, sentimentClient, analyze.NewGrammarClient(), analyze.NewBingClient(os.Getenv("BING_KEY_1")))
		break
	default:
		fmt.Printf("%s is not a valid analysis type; available types are grammar, sentiment and all", analysisType)
		os.Exit(1)
	}

	issues, err := boltDB.Issues()
	if err != nil {
		log.Fatalf("could not retrieve issues from Bolt DB: %v\n", err)
	}

	scoreMap, err := analyze.MultipleScores(issues[:31], clients...)
	if err != nil {
		log.Fatalf("could not calculate scores: %v\n", err)
	}

	for k, v := range scoreMap {
		switch k {
		case "GRAMMAR":
			for i := range v {
				issues[i].GrammarErrCount = v[i]
			}
			break
		case "SENTIMENT":
			for i := range v {
				issues[i].SentimentScore = v[i]
			}
			break
		}
	}

	err = boltDB.InsertIssues(issues...)
	if err != nil {
		log.Fatalf("could not insert issues back into db: %v\n", err)
	}
}
