package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nclandrei/ticketguru/analyze"
	"github.com/nclandrei/ticketguru/db"
	// "github.com/nclandrei/ticketguru/jira"
	"log"
	"os"
)

func main() {
	boltDB, err := db.NewBoltDB("users.db")
	if err != nil {
		log.Fatalf("could not access Bolt DB: %v\n", err)
	}

	var analysisType string
	flag.StringVar(&analysisType, "type", "all", "type of analysis to run; available types: grammar, sentiment, " +
		"stack_traces, steps_to_reproduce, attachments, comment_complexity, fields_complexity, all")

	flag.Parse()

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env file: %v\n", err)
	}

	var clients []analyze.Scorer
	var analysisFuncs []analyze.TicketAnalysis

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
	case "steps_to_reproduce":
		analysisFuncs = append(analysisFuncs, analyze.HaveStepsToReproduce)
		break
	case "stack_traces":
		analysisFuncs = append(analysisFuncs, analyze.HaveStackTrace)
		break
	case "have"
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

	// totalIssueLen, err := boltDB.Size()
	// if err != nil {
	// 	log.Fatalf("could not retrieve issues bucket size: %v\n", err)
	// }

	// sliceSize := 10000
	// highBound := sliceSize
	// issues := make([]jira.Ticket, sliceSize)

	// for i := 0; i < totalIssueLen; i += sliceSize {
	// 	if i+highBound > totalIssueLen {
	// 		highBound = totalIssueLen % sliceSize
	// 	}
	// 	issues, err = boltDB.Slice(i, i+highBound)
	// 	if err != nil {
	// 		log.Fatalf("could not get issue slice: %v\n", err)
	// 	}
	// 	err = analyze.MultipleScores(issues, clients...)
	// 	if err != nil {
	// 		log.Printf("could not calculate scores: \n%v\n", err)
	// 	}
	// 	err = boltDB.Insert(issues...)
	// 	if err != nil {
	// 		log.Fatalf("could not insert issues in db: %v\n", err)
	// 	}
	// }
	tickets, err := boltDB.Tickets()
	if err != nil {
		log.Fatalf("could not get all issues inside the database: %v\n", err)
	}
	analyze.CountWordsComments(tickets...)
	analyze.CountWordsSummaryDesc(tickets...)
	analyze.HaveStackTrace(tickets...)
	analyze.HaveStepsToReproduce(tickets...)
	analyze.TimesToClose(tickets...)
	err = boltDB.Insert(tickets...)
	if err != nil {
		log.Fatalf("could not insert tickets: %v\n", err)
	}
}
