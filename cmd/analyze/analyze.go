package main

import (
	"context"
	"log"

	"github.com/nclandrei/L5-Project/gcp"

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

	langClient, err := gcp.NewLanguageClient(context.Background())
	if err != nil {
		log.Fatalf("could not create new language client: %v\n", err)
	}

	ii, err := boltDB.Issues()
	if err != nil {
		log.Fatalf("could not retrieve issues: %v\n", err)
	}

	for _, issue := range ii {
		bIssue, err := boltDB.IssueByKey(issue.Key)
		if err != nil {
			log.Fatalf("could not retrieve issue {%s} from bolt: %v\n", err)
		}
		if bIssue.SentimentScore == nil {
			commentScorelangClient.SentimentScoreFromText
		}
	}

	withAttch, withoutAttch := analyze.AttachmentsAnalysis(ii)
	wordCountSlice, timeDiffs := analyze.WordinessAnalysis(ii, "description")

	err = plotter.DrawAttachmentsBarchart("Attachments Analysis", "Time-To-Resolve", withAttch, withoutAttch)
	if err != nil {
		log.Fatalf("could not draw attachments barchart: %v\n", err)
	}

	err = plotter.DrawPlot("Description Analysis", "#Words", "Time-To-Resolve", wordCountSlice, timeDiffs)
	if err != nil {
		log.Fatalf("could not draw comment plot: %v\n", err)
	}
}
