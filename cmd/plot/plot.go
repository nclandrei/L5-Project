package main

import (
	"flag"
	"log"

	"github.com/nclandrei/ticketguru/plot"

	"github.com/nclandrei/ticketguru/analyze"

	"github.com/nclandrei/ticketguru/db"
)

var (
	dbPath = flag.String(
		"dbPath",
		"/Users/nclandrei/Code/go/src/github.com/nclandrei/ticketguru/users.db",
		"path to Bolt database file",
	)
	plotType = flag.String("type", "all", "plot to draw")
)

func main() {
	boltDB, err := db.NewBoltDB(*dbPath)
	if err != nil {
		log.Fatalf("could not retrieve database: %v\n", err)
	}

	plotter, err := plot.NewPlotter()
	if err != nil {
		log.Fatalf("could not create new plotter: %v\n", err)
	}

	ii, err := boltDB.Issues()
	if err != nil {
		log.Fatalf("could not retrieve issues: %v\n", err)
	}

	withAttch, withoutAttch := analyze.AttachmentsAnalysis(ii)
	err = plotter.DrawAttachmentsBarchart("Attachments Analysis", "Time-To-Resolve", withAttch, withoutAttch)
	if err != nil {
		log.Fatalf("could not draw attachments barchart: %v\n", err)
	}

	wordCountSlice, timeDiffs := analyze.WordinessAnalysis(ii, "description")
	err = plotter.DrawPlot("Description Analysis", "#Words", "Time-To-Resolve", wordCountSlice, timeDiffs)
	if err != nil {
		log.Fatalf("could not draw comment plot: %v\n", err)
	}

	sentimentScores, timeDiffs := analyze.SentimentScoreAnalysis(ii)
	err = plotter.DrawPlot("Sentiment Score Analysis", "Score", "Time-To-Resolve", sentimentScores, timeDiffs)
	if err != nil {
		log.Fatalf("could not draw sentiment score plot: %v\n", err)
	}
}
