package gcp

import (
	"context"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

// LangClient defines a GCP Language Client
type LangClient struct {
	*language.Client
	ctx context.Context
}

// NewLanguageClient returns a new language clients alogn with its context
func NewLanguageClient(ctx context.Context) (*LangClient, error) {
	client, err := language.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &LangClient{
		Client: client,
		ctx:    ctx,
	}, nil
}

// CommSentimentScore calculates the sentiment score for an issue's comments after querying GCP.
func (client *LangClient) CommSentimentScore(concatComments string) (float32, error) {
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
