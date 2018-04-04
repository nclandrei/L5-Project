package main

import (
	"context"
	"github.com/nclandrei/L5-Project/analyze"
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

	doneCh := make(chan struct{})

	grammarClient := language.NewGrammarClient()

	sentimentClient, err := language.NewSentimentClient(context.Background())
	if err != nil {
		log.Fatalf("could not create GCP sentiment client: %v\n", err)
	}

	go func() {
		scores, err := grammarClient.Scores(issues...)
		if err != nil {
			log.Printf("could not retrieve grammar scores for issues: %v\n", err)
		}
		for i := range scores {
			issues[i].GrammarErrCount = scores[i]
		}
		doneCh <- struct{}{}
	}()

	for _, issue := range issues {
		if issue.SentimentScore != 0 {
			continue
		}
		concatComm, err := analyze.ConcatenateComments(issue)
		if err != nil {
			log.Printf("could not concatenate comments for issue {%s}: %v\n", issue.Key, err)
			continue
		}
		score, err := sentimentClient.CommentScore(concatComm)
		if err != nil {
			log.Printf("could not calculate sentiment score for issue {%s}: %v\n", issue.Key, err)
			continue
		}
		issue.SentimentScore = score
	}

	<-doneCh

	err = boltDB.InsertIssues(issues...)
	if err != nil {
		log.Fatalf("could not store issues inside database: %v\n", err)
	}
}
