package main

import (
	"flag"
	"github.com/nclandrei/ticketguru/db"
	"github.com/nclandrei/ticketguru/plot"
	"log"
	"sync"
)

var (
	dbPath = flag.String(
		"dbPath",
		"/Users/nclandrei/Code/go/src/github.com/nclandrei/ticketguru/users.db",
		"path to Bolt database file",
	)
	pType = flag.String("type", "all", "plot(s) to draw - available types: grammar, sentiment, steps_to_reprodce"+
		"stack_traces, attachments, comments_complexity, fields_complexity, all")
)

func main() {
	boltDB, err := db.NewBolt(*dbPath)
	if err != nil {
		log.Fatalf("could not open bolt db: %v\n", err)
	}
	tickets, err := boltDB.Tickets()
	if err != nil {
		log.Fatalf("could not get tickets from bolt db: %v\n", err)
	}
	var funcs []plot.Plot
	switch *pType {
	case "grammar":
		funcs = append(funcs, plot.GrammarCorrectness)
		break
	case "sentiment":
		funcs = append(funcs, plot.SentimentAnalysis)
		break
	case "steps_to_reproduce":
		funcs = append(funcs, plot.StepsToReproduce)
		break
	case "stack_traces":
		funcs = append(funcs, plot.Stacktraces)
		break
	case "attachments":
		funcs = append(funcs, plot.Attachments)
		break
	case "comments_complexity":
		funcs = append(funcs, plot.CommentsComplexity)
		break
	case "fields_complexity":
		funcs = append(funcs, plot.FieldsComplexity)
		break
	case "all":
		funcs = append(funcs, plot.CommentsComplexity, plot.FieldsComplexity, plot.SentimentAnalysis, plot.GrammarCorrectness,
			plot.Stacktraces, plot.StepsToReproduce)
		break
	}
	var wg sync.WaitGroup
	for _, f := range funcs {
		wg.Add(1)
		go func(f plot.Plot) {
			defer wg.Done()
			err := f(tickets...)
			if err != nil {
				log.Printf("could not plot data: %v\n", err)
			}
		}(f)
	}
	wg.Wait()
}
