package analyze

import (
	"context"
	"encoding/json"
	"github.com/nclandrei/L5-Project/jira"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

const (
	languageToolRateLimit = 20                                      // defines number of requests permitted per minute
	gcpRateLimit          = 600                                     // defines the GCP Natural Language API rate limit per minute
	bingRateLimit         = 100                                     // defines Bing Spell Check API rate limit per second
	languageToolAPIPath   = "https://languagetool.org/api/v2/check" // URL path to LanguageTool API
	bingAPIPath           = "https://api.cognitive.microsoft.com/bing/v7.0/SpellCheck"
)

// Scorer defines an interface for holding the different types of language scorers available.
type Scorer interface {
	Scores(...jira.Issue) ([]float64, error)
	Name() string
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
	return &BingClient{
		Client: &http.Client{},
		key:    key,
	}
}

// Name returns the name of the Bing client.
func (client *BingClient) Name() string {
	return "GRAMMAR"
}

// Scores returns the grammar correctness scores for all issues given as input parameters.
func (client *BingClient) Scores(issues ...jira.Issue) ([]float64, error) {
	var scores []float64
	errCh := make(chan error, len(issues))
	var rateLimit int
	if bingRateLimit > len(issues) {
		rateLimit = len(issues)
	} else {
		rateLimit = bingRateLimit
	}
	for i := 0; i < len(issues); i += rateLimit {
		for _, issue := range issues[i:(i + rateLimit)] {
			go func(issue jira.Issue) {
				strToAnalyze, err := concatAndRemoveNewlines(issue.Fields.Summary, issue.Fields.Description)
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
				scores = append(scores, float64(len(bingResponse.FlaggedTokens)))
				errCh <- nil
			}(issue)
		}
		time.Sleep(1 * time.Second)
	}
	for i := 0; i < len(issues); i++ {
		if err := <-errCh; err != nil {
			return scores, err
		}
	}
	return scores, nil
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

// Name returns the name of the GCP Natural Language client.
func (client SentimentClient) Name() string {
	return "SENTIMENT"
}

// Scores calculates the sentiment score for an issue's comments after querying GCP.
func (client *SentimentClient) Scores(issues ...jira.Issue) ([]float64, error) {
	scores := make([]float64, len(issues))
	errCh := make(chan error, len(issues))
	var rateLimit int
	if gcpRateLimit > len(issues) {
		rateLimit = len(issues)
	} else {
		rateLimit = gcpRateLimit
	}
	for i := 0; i < len(issues); i += rateLimit {
		for _, issue := range issues[i:(i + rateLimit)] {
			go func(issue jira.Issue) {
				if issue.SentimentScore != 0 {
					return
				}
				concatComm, err := concatenateComments(issue)
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
				scores = append(scores, float64(sentiment.DocumentSentiment.Score))
				errCh <- nil
			}(issue)
		}
		time.Sleep(1 * time.Minute)
	}
	for i := 0; i < len(issues); i++ {
		if err := <-errCh; err != nil {
			return scores, err
		}
	}
	return scores, nil
}

// MultipleScores takes multiple issues and scorers and returns a map for each scorer to its corresponding scores.
func MultipleScores(issues []jira.Issue, scorers ...Scorer) (map[string][]float64, error) {
	scoreMap := make(map[string][]float64)
	errCh := make(chan error, len(scorers))
	for i := range scorers {
		go func(index int) {
			scores, err := scorers[index].Scores(issues...)
			if err != nil {
				errCh <- err
			} else {
				scoreMap[scorers[index].Name()] = scores
				errCh <- nil
			}
		}(i)
	}
	for i := 0; i < len(scorers); i++ {
		if err := <-errCh; err != nil {
			return scoreMap, err
		}
	}
	return scoreMap, nil
}
