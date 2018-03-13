package analyze

import (
	"context"
	"strings"
	"time"

	language "cloud.google.com/go/language/apiv1"
	"github.com/nclandrei/L5-Project/jira"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

// WordinessAnalysis returns wordiness of a field (summary/comment/description) and time-to-complete (in hours)
func WordinessAnalysis(issues []jira.Issue, field string) ([]float64, []float64) {
	var wordCountSlice []float64
	var timeDiffs []float64
	for _, issue := range issues {
		timeDiff := timeToResolve(issue)
		if timeDiff > -1 && isIssueHighPriority(issue) {
			switch field {
			case "description":
				wordCountSlice = append(wordCountSlice, float64(calculateNumberOfWords(issue.Fields.Description)))
				break
			case "summary":
				wordCountSlice = append(wordCountSlice, float64(calculateNumberOfWords(issue.Fields.Summary)))
				break
			case "comment":
				wc := 0
				for _, comment := range issue.Fields.Comments.Comments {
					wc += calculateNumberOfWords(comment.Body)
				}
				wordCountSlice = append(wordCountSlice, float64(wc))
				break
			}
			timeDiffs = append(timeDiffs, timeDiff)
		}
	}
	return wordCountSlice, timeDiffs
}

// AttachmentsAnalysis returns time-to-complete (in hours) for all issues with and without attachments
func AttachmentsAnalysis(issues []jira.Issue) ([]float64, []float64) {
	var withAttchTimeDiffs []float64
	var withoutAttchTimeDiffs []float64
	for _, issue := range issues {
		timeDiff := timeToResolve(issue)
		if timeDiff > -1 && isIssueHighPriority(issue) {
			if len(issue.Fields.Attachments) > 0 {
				withAttchTimeDiffs = append(withAttchTimeDiffs, timeDiff)
			} else {
				withoutAttchTimeDiffs = append(withoutAttchTimeDiffs, timeDiff)
			}
		}
	}
	return withAttchTimeDiffs, withoutAttchTimeDiffs
}

// SentimentScoreFromText calculates the sentiment score for a fragment after querying GCP
func SentimentScoreFromText(doc string) (float32, error) {
	ctx := context.Background()

	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		return 0, err
	}

	sentiment, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: doc,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})
	if err != nil {
		return 0, err
	}

	return sentiment.DocumentSentiment.Score, nil
}

// calculateNumberOfWords returns the number of words in a string
func calculateNumberOfWords(s string) int {
	wordCount := 0
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		wordCount += len(strings.Split(strings.TrimSpace(line), " "))
	}
	return wordCount
}

// GetAttachmentType returns the attachment type based on the file extension
func getAttachmentType(filename string) jira.AttachmentType {
	extIndex := strings.LastIndex(filename, ".")
	ext := filename[(extIndex + 1):]
	switch ext {
	case "md":
		return jira.Text
	case "txt":
		return jira.Text
	case "pdf":
		return jira.Text
	case "png":
		return jira.Image
	case "jpg":
		return jira.Image
	case "jpeg":
		return jira.Image
	case "gif":
		return jira.Image
	case "bmp":
		return jira.Image
	case "mp4":
		return jira.Video
	case "avi":
		return jira.Video
	case "mkv":
		return jira.Video
	default:
		return jira.Code
	}
}

// calculateJTimeDifference calculates the duration in hours between 2 different timestamps
func calculateJTimeDifference(t1, t2 jira.JTime) float64 {
	return time.Time(t1).Sub(time.Time(t2)).Hours()
}

func isIssueHighPriority(issue jira.Issue) bool {
	return issue.Fields.Priority.ID == "1" || issue.Fields.Priority.ID == "2"
}

func timeToResolve(issue jira.Issue) float64 {
	for _, history := range issue.Changelog.Histories {
		for _, item := range history.Items {
			if item.Field == "status" && item.FromString == "Open" && item.ToString == "Closed" {
				return calculateJTimeDifference(history.Created, issue.Fields.Created)
			}
		}
	}
	return -1
}
