package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getIssues(
	responses chan<- []byte,
	done chan<- bool,
	paginationIndex int,
	pageCount int,
	projectName string) {

	fmt.Println(paginationIndex)

	requestBody := &JqlRequestBody{
		Jql:        fmt.Sprintf("project=%s", projectName),
		StartAt:    paginationIndex * pageCount,
		MaxResults: pageCount,
	}

	req, _ := json.Marshal(requestBody)

	resp, err := http.Post(jiraURL, "application/json", bytes.NewBuffer(req))
	if err != nil {
		fmt.Printf("Could not send request: %v", err)
	} else {
		fmt.Println("response Status:", resp.Status)
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			if bodyJSON, err := json.Marshal(bodyBytes); err != nil {
				fmt.Printf("Could not marshal response to JSON: %v", err)
			} else {
				responses <- bodyJSON
			}
		}
	}
	done <- true
}
