package main

import (
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
	_, err = client.Scores(issues[:41]...)
	if err != nil {
		log.Fatalf("could not retrieve grammar scores for issues: %v\n", err)
	}
}
