package language

import (
	"context"
	"github.com/nclandrei/L5-Project/analyze"
	"github.com/nclandrei/L5-Project/jira"
	"time"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

const (
	apiRateLimit = 600 // defines the GCP Natural Language API rate limit per minute
)

// SentimentClient defines a GCP Language Client
type SentimentClient struct {
	*language.Client
	ctx context.Context
}

// NewSentimentClient returns a new language clients alogn with its context
func NewSentimentClient(ctx context.Context) (*SentimentClient, error) {
	client, err := language.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &SentimentClient{
		Client: client,
		ctx:    ctx,
	}, nil
}

// SentimentScores calculates the sentiment score for an issue's comments after querying GCP.
func (client *SentimentClient) SentimentScores(issues ...jira.Issue) ([]float64, error) {
	scores := make([]float64, len(issues))
	for i := 0; i < len(issues); i += apiRateLimit {
		for _, issue := range issues[i:(i + 20)] {
			concatComm, err := analyze.ConcatenateComments(issue)
			if err != nil {
				return scores, err
			}
			sentiment, err := client.AnalyzeSentiment(client.ctx, &languagepb.AnalyzeSentimentRequest{
				Document: &languagepb.Document{
					Source: &languagepb.Document_Content{
						Content: concatComm,
					},
					Type: languagepb.Document_PLAIN_TEXT,
				},
				EncodingType: languagepb.EncodingType_UTF8,
			})
			if err != nil {
				return scores, err
			}
			scores = append(scores, float64(sentiment.DocumentSentiment.Score))
		}
		time.Sleep(1 * time.Minute)
	}
	return scores, nil
}
