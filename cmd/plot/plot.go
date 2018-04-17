package main

import (
	"flag"
	"github.com/nclandrei/ticketguru/db"
	"github.com/nclandrei/ticketguru/plot"
	"log"
)

var (
	dbPath = flag.String(
		"dbPath",
		"/Users/nclandrei/Code/go/src/github.com/nclandrei/ticketguru/users.db",
		"path to Bolt database file",
	)
	pType = flag.String("type", "all", "plot(s) to draw")
)

func main() {
	boltDB, err := db.NewBolt(*dbPath)
	if err != nil {
		log.Fatalf("could not open bolt db: %v\n", err)
	}
	tickets, err := boltDB.Slice(0, 1000)
	if err != nil {
		log.Fatalf("could not get ticket slice from bolt db: %v\n", err)
	}
	var funcs []plot.Plot
	switch *pType {
	case "all":
		funcs = append(funcs, plot.CommentsComplexity, plot.FieldsComplexity)
		break
	}
	for _, f := range funcs {
		err := f(tickets...)
		if err != nil {
			log.Fatalf("could not plot data: %v\n", err)
		}
	}
}
