package language

import (
	"context"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
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

// CommentScore calculates the sentiment score for an issue's comments after querying GCP.
func (client *SentimentClient) CommentScore(concatComments string) (float32, error) {
	sentiment, err := client.AnalyzeSentiment(client.ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: concatComments,
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
