package main

import (
	"context"
	"github.com/nclandrei/L5-Project/db"
	"github.com/nclandrei/L5-Project/language"
	"log"
)

func main() {
	boltDB, err := db.NewBoltDB("users.db")
	if err != nil {
		log.Fatalf("could not access Bolt DB: %v\n", err)
	}

	issues, err := boltDB.Issues()
	if err != nil {
		log.Fatalf("could not retrieve issues from Bolt DB: %v\n", err)
	}

	grammarClient := language.NewGrammarClient()

	sentimentClient, err := language.NewSentimentClient(context.Background())
	if err != nil {
		log.Fatalf("could not create GCP sentiment client: %v\n", err)
	}

	scoreMap, err := language.MultipleScores(issues, grammarClient, sentimentClient)
	if err != nil {
		log.Fatalf("could not calculate scores: %v\n", err)
	}
	log.Println(scoreMap)
}
