package main

import (
	"log"

	"github.com/nclandrei/L5-Project/db"
)

func main() {
	boltDB, err := db.NewBoltDB("/Users/nclandrei/Code/go/src/github.com/nclandrei/L5-Project/users.db")
	if err != nil {
		log.Fatalf("could not create Bolt DB: %v\n", err)
	}

	if err != nil {
		log.Fatalf("could not insert issues: %v\n", err)
	}

	dbIssues, err := boltDB.GetIssues()
	if err != nil {
		log.Fatalf("could not retrieve issues: %v\n", err)
	}

	log.Println(dbIssues)
}
