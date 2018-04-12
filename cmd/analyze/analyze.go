package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
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

	var analysisType string
	flag.StringVar(&analysisType, "type", "all", "type of analysis to run; available types: langTool,"+
		" sentiment, bing, all (sentiment, langTool, bing spell check)")

	flag.Parse()

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env file: %v\n", err)
	}

	clients := make([]analyze.Scorer, 3)

	switch analysisType {
	case "langTool":
		clients = append(clients, analyze.NewLanguageToolClient())
		break
	case "bing":
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
			analyze.NewLanguageToolClient(),
			analyze.NewBingClient(os.Getenv("BING_KEY_1")),
		)
		break
	default:
		fmt.Printf("%s is not a valid analysis type; available types are grammar, sentiment and all", analysisType)
		os.Exit(1)
	}

	issues, err := boltDB.Issues()
	if err != nil {
		log.Fatalf("could not retrieve issues from Bolt DB: %v\n", err)
	}

	scoreMap, err := analyze.MultipleScores(issues[:10], clients...)
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
		case "SPELL_CHECK":
			for i := range v {
				issues[i].GrammarErrCount = v[i]
			}
		}

		err = boltDB.InsertIssues(issues...)
		if err != nil {
			log.Fatalf("could not insert issues back into db: %v\n", err)
		}
	}
}
