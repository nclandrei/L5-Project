package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	// numberOfIssues := flag.Int("issuesCount", 50000, "defines the number of issues to be retrieved")

	flag.Parse()

	responses := make(chan Issue)
	done := make(chan bool)
	var respSlice []Issue

	jiraClient := http.DefaultClient

	for i := 0; i < 1; i++ {
		go getIssues(responses, done, i, 1, *projectName, jiraClient)
	}

	doneCounter := 0

	for doneCounter < 1 {
		select {
		case newResponse := <-responses:
			respSlice = append(respSlice, newResponse)
		case <-done:
			doneCounter++
		}
	}

	for _, value := range respSlice {
		log.Println(value)
	}
}
