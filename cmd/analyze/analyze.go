package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nclandrei/ticketguru/analyze"
	"github.com/nclandrei/ticketguru/db"
	"sync"
	// "github.com/nclandrei/ticketguru/jira"
	"log"
	"os"
)

func main() {
	boltDB, err := db.NewBolt("users.db")
	if err != nil {
		log.Fatalf("could not access Bolt DB: %v\n", err)
	}

	var analysisType string
	flag.StringVar(&analysisType, "type", "all", "type of analysis to run; available types: grammar, sentiment, "+
		"stack_traces, steps_to_reproduce, attachments, comment_complexity, fields_complexity, all")

	flag.Parse()

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env file: %v\n", err)
	}

	var clients []analyze.Scorer
	var analysisFuncs []analyze.TicketAnalysis
	analysisFuncs = append(analysisFuncs, analyze.TimesToClose)

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
		analysisFuncs = append(analysisFuncs, analyze.StepsToReproduce)
		break
	case "stack_traces":
		analysisFuncs = append(analysisFuncs, analyze.StackTraces)
		break
	case "attachments":
		analysisFuncs = append(analysisFuncs, analyze.Attachments)
		break
	case "comment_complexity":
		analysisFuncs = append(analysisFuncs, analyze.CommentsComplexity)
		break
	case "fields_complexity":
		analysisFuncs = append(analysisFuncs, analyze.FieldsComplexity)
		break
	case "all":
		analysisFuncs = append(analysisFuncs, analyze.StepsToReproduce, analyze.StackTraces, analyze.Attachments,
			analyze.CommentsComplexity, analyze.FieldsComplexity)
		break
	default:
		fmt.Printf("%s is not a valid analysis type; available types are grammar, sentiment and all", analysisType)
		os.Exit(1)
	}

	tickets, err := boltDB.Tickets()
	if err != nil {
		log.Fatalf("could not get all issues inside the database: %v\n", err)
	}

	analyze.MultipleScores(tickets, clients...)

	var wg sync.WaitGroup
	for _, f := range analysisFuncs {
		wg.Add(1)
		go func(f analyze.TicketAnalysis) {
			defer wg.Done()
			f(tickets...)
		}(f)
	}

	wg.Done()

	err = boltDB.Insert(tickets...)
	if err != nil {
		log.Fatalf("could not insert tickets: %v\n", err)
	}
}
