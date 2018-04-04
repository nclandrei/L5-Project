package main

import (
	"fmt"
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
	client := language.NewGrammarClient()
	scores, err := client.Scores(issues[:41]...)
	if err != nil {
		log.Printf("could not retrieve grammar scores for issues: %v\n", err)
	}
	for i := range scores {
		issues[i].GrammarErrCount = scores[i]
		fmt.Println(issues[i].GrammarErrCount)
	}
	err = boltDB.InsertIssues(issues...)
	if err != nil {
		log.Fatalf("could not store issues inside database: %v\n", err)
	}
}
