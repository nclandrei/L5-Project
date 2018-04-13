package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nclandrei/L5-Project/jira"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

const (
	gcpRateLimit  = 600 // defines the GCP Natural Language API rate limit per minute
	bingRateLimit = 100 // defines Bing Spell Check API rate limit per second
	bingAPIPath   = "https://api.cognitive.microsoft.com/bing/v7.0/SpellCheck"
)

// Scorer defines an interface for holding the different types of language scorers available.
type Scorer interface {
	Scores(...jira.Issue) error
}

// BingClient defines a new Bing Spell Check client.
type BingClient struct {
	*http.Client
	key string
}

// BingResponse holds responses retrieved from Bing Spell Check API.
type BingResponse struct {
	Type          string `json:"-"`
	FlaggedTokens []BingFlaggedToken
}

// BingFlaggedToken holds information regarding flagged tokens inside the text passed in the request.
type BingFlaggedToken struct {
	Offset int    `json:"offset"`
	Token  string `json:"token"`
	Type   string `json:"type"`
}

// NewBingClient returns a new Bing Spell Check API client.
func NewBingClient(key string) *BingClient {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}
	client := &http.Client{
		Timeout:   90 * time.Second,
		Transport: transport,
	}
	return &BingClient{
		Client: client,
		key:    key,
	}
}

// Scores returns the grammar correctness scores for all issues given as input parameters.
func (client *BingClient) Scores(issues ...jira.Issue) error {
	errCh := make(chan error, len(issues))
	var rateLimit int
	if bingRateLimit > len(issues) {
		rateLimit = len(issues)
	} else {
		rateLimit = bingRateLimit
	}
	for i := 0; i < len(issues); i += rateLimit {
		for j := range issues[i:(i + rateLimit)] {
			go func(i, j int) {
				if issues[i+j].GrammarCorrectness.HasScore {
					errCh <- nil
					return
				}
				strToAnalyze, err := concatAndRemoveNewlines(issues[i+j].Fields.Summary, issues[i+j].Fields.Description)
				if err != nil {
					errCh <- err
					return
				}
				values := url.Values{}
				values.Set("Text", strToAnalyze)
				req, err := http.NewRequest(
					"POST",
					bingAPIPath,
					strings.NewReader(values.Encode()),
				)
				if err != nil {
					errCh <- err
					return
				}
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("Ocp-Apim-Subscription-Key", client.key)
				resp, err := client.Do(req)
				if err != nil {
					errCh <- err
					return
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					errCh <- err
					return
				}
				bingResponse := &BingResponse{}
				err = json.Unmarshal(body, bingResponse)
				if err != nil {
					errCh <- err
					return
				}
				issues[i+j].GrammarCorrectness.Score = len(bingResponse.FlaggedTokens)
				issues[i+j].GrammarCorrectness.HasScore = true
				errCh <- nil
			}(i, j)
		}
		time.Sleep(1 * time.Second)
	}
	var strBuilder strings.Builder
	for i := 0; i < len(issues); i++ {
		if err := <-errCh; err != nil {
			strBuilder.WriteString("error while retrieving grammar scores: ")
			strBuilder.WriteString(err.Error())
			strBuilder.WriteRune('\n')
		}
	}
	if strBuilder.Len() > 0 {
		return fmt.Errorf(strBuilder.String())
	}
	return nil
}

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

// Scores calculates the sentiment score for an issue's comments after querying GCP.
func (client *SentimentClient) Scores(issues ...jira.Issue) error {
	errCh := make(chan error, len(issues))
	var rateLimit int
	if gcpRateLimit > len(issues) {
		rateLimit = len(issues)
	} else {
		rateLimit = gcpRateLimit
	}
	for i := 0; i < len(issues); i += rateLimit {
		for j := range issues[i:(i + rateLimit)] {
			go func(i, j int) {
				if issues[i+j].Sentiment.HasScore {
					errCh <- nil
					return
				}
				concatComm, err := concatenateComments(issues[i+j])
				if err != nil {
					errCh <- err
					return
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
					errCh <- err
					return
				}
				issues[i+j].Sentiment.HasScore = true
				issues[i+j].Sentiment.Score = float64(sentiment.DocumentSentiment.Score)
				errCh <- nil
			}(i, j)
		}
		time.Sleep(1 * time.Minute)
	}
	var strBuilder strings.Builder
	for i := 0; i < len(issues); i++ {
		if err := <-errCh; err != nil {
			strBuilder.WriteString("error while retrieving sentiment scores: ")
			strBuilder.WriteString(err.Error())
			strBuilder.WriteRune('\n')
		}
	}
	if strBuilder.Len() > 0 {
		return fmt.Errorf(strBuilder.String())
	}
	return nil
}

// MultipleScores takes multiple issues and scorers and returns a map for each scorer to its corresponding scores.
func MultipleScores(issues []jira.Issue, scorers ...Scorer) error {
	errCh := make(chan error, len(scorers))
	for i := range scorers {
		go func(i int) {
			errCh <- scorers[i].Scores(issues...)
		}(i)
	}
	for i := 0; i < len(scorers); i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}
