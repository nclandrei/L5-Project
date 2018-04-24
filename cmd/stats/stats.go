package main

import (
	"flag"
	"github.com/nclandrei/ticketguru/db"
	"github.com/nclandrei/ticketguru/stats"
	"log"
	"sync"
)

var (
	dbPath = flag.String(
		"dbPath",
		"/Users/nclandrei/Code/go/src/github.com/nclandrei/ticketguru/issues.db",
		"path to Bolt database file",
	)
)

func main() {
	boltDB, err := db.NewBolt(*dbPath)
	if err != nil {
		log.Fatalf("could not access Bolt DB: %v\n", err)
	}

	var analysisType string
	flag.StringVar(&analysisType, "type", "all", "type of statistics to run; available types: grammar, sentiment, "+
		"stack_traces, steps_to_reproduce, attachments, comment_complexity, fields_complexity, all")

	flag.Parse()

	categoricalTests := map[string]stats.CategoricalTest{
		"Attachments":        stats.Attachments,
		"Steps To Reproduce": stats.StepsToReproduce,
		"Stack Traces":       stats.Stacktraces,
	}
	continuousTests := map[string]stats.ContinuousTest{
		"Comments Complexity": stats.CommentsComplexity,
		"Fields Complexity":   stats.FieldsComplexity,
		"Sentiment Analysis":  stats.Sentiment,
		"Grammar Correctness": stats.Grammar,
	}

	tickets, err := boltDB.Tickets()
	if err != nil {
		log.Fatalf("could not fetch tickets from bolt db: %v\n", err)
	}

	var wg sync.WaitGroup
	for k, v := range categoricalTests {
		wg.Add(1)
		go func(name string, f stats.CategoricalTest) {
			defer wg.Done()
			result, err := f(tickets...)
			if err != nil {
				log.Printf("could not compute statistical test: %v\n", err)
				return
			}
			log.Printf("%s --- P: %f --- mean_1: %f --- mean_2: %f\n", name, result.P, result.N1Mean, result.N2Mean)
		}(k, v)
	}

	for k, v := range continuousTests {
		wg.Add(1)
		go func(name string, f stats.ContinuousTest) {
			defer wg.Done()
			result := f(tickets...)
			if err != nil {
				log.Printf("could not compute statistical test: %v\n", err)
			}
			log.Printf("%s --- Rs: %f --- P: %f\n", name, result.Rs, result.P)
		}(k, v)
	}

	wg.Wait()
}
