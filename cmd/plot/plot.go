package main

import (
	"flag"
)

var (
	dbPath = flag.String(
		"dbPath",
		"/Users/nclandrei/Code/go/src/github.com/nclandrei/ticketguru/users.db",
		"path to Bolt database file",
	)
	plotType = flag.String("type", "all", "plot to draw")
)

func main() {}
