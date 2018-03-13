package main

import (
	"log"

	"github.com/nclandrei/L5-Project/plot"

	"github.com/nclandrei/L5-Project/analyze"

	"github.com/nclandrei/L5-Project/db"
)

// var (
// 	analysisTask := flag.String("task", "attachments", "")
// )

func main() {
	boltDB, err := db.NewBoltDB("/Users/nclandrei/Code/go/src/github.com/nclandrei/L5-Project/users.db")
	if err != nil {
		log.Fatalf("could not create Bolt DB: %v\n", err)
	}

	plotter, err := plot.NewPlotter()
	if err != nil {
		log.Fatalf("could not create new plotter: %v\n", err)
	}

	dbIssues, err := boltDB.GetIssues()
	if err != nil {
		log.Fatalf("could not retrieve issues: %v\n", err)
	}

	withAttch, withoutAttch := analyze.AttachmentsAnalysis(dbIssues)
	wordCountSlice, timeDiffs := analyze.WordinessAnalysis(dbIssues, "description")

	err = plotter.DrawAttachmentsBarchart("Attachments Analysis", "Time-To-Resolve", withAttch, withoutAttch)
	if err != nil {
		log.Fatalf("could not draw attachments barchart: %v\n", err)
	}

	err = plotter.DrawPlot("Description Analysis", "#Words", "Time-To-Resolve", wordCountSlice, timeDiffs)
	if err != nil {
		log.Fatalf("could not draw comment plot: %v\n", err)
	}
}
