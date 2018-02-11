package main

import (
	"flag"
	"log"
)

func main() {
	projectName := flag.String("project", "Kafka", "defines the name of the project to be queried upon")
	// numberOfIssues := flag.Int("issuesCount", 50000, "defines the number of issues to be retrieved")

	flag.Parse()

	responses := make(chan Fields)
	done := make(chan bool)
	var respSlice []Fields

	jiraClient := NewJiraClient()

	for i := 0; i < 1; i++ {
		go jiraClient.GetPaginatedIssues(responses, done, i, 1, *projectName)
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
