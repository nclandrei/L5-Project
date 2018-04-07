package language

import (
	"context"
	"encoding/json"
	"github.com/nclandrei/L5-Project/analyze"
	"github.com/nclandrei/L5-Project/jira"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

const (
	languageToolRateLimit = 20                                      // defines number of requests permitted per minute
	gcpRateLimit          = 600                                     // defines the GCP Natural Language API rate limit per minute
	languageToolAPIPath   = "https://languagetool.org/api/v2/check" // URL path to LanguageTool API
)

// Scorer defines an interface for holding the different types of language scorers available.
type Scorer interface {
	Scores(...jira.Issue) ([]float64, error)
	Name() string
}

// GrammarClient defines the LanguageTool http client.
type GrammarClient struct {
	*http.Client
	rateLimit int
	path      string
}

// GrammarResponse defines the response retrieved via LanguageTool API.
type GrammarResponse struct {
	Matches []GrammarMatch `json:"matches"`
}

// GrammarMatch defines a match for an issue found in the parsed text.
type GrammarMatch struct {
	Rule GrammarRule `json:"rule"`
}

// GrammarRule defines all the necessary info needed to understand a grammar error from LanguageTool.
type GrammarRule struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	IssueType   string `json:"issueType"`
	Category    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}

// NewGrammarClient returns a new Grammar client.
func NewGrammarClient() *GrammarClient {
	return &GrammarClient{
		Client:    http.DefaultClient,
		rateLimit: languageToolRateLimit,
		path:      languageToolAPIPath,
	}
}

// newRequestBody returns a request body for a LanguageTool API call.
func newRequestBody(text string) io.Reader {
	data := url.Values{}
	data.Set("language", "en")
	data.Set("text", text)
	return strings.NewReader(data.Encode())
}

// Name returns the name of the grammar scorer.
func (client GrammarClient) Name() string {
	return "GRAMMAR"
}

// Scores returns the grammar scores for all issues passed as arguments.
func (client *GrammarClient) Scores(issues ...jira.Issue) ([]float64, error) {
	var scores []float64
	for i := 0; i < len(issues); i += languageToolRateLimit {
		for _, issue := range issues[i:(i + languageToolRateLimit)] {
			strToAnalyze := strings.Join([]string{issue.Fields.Summary, issue.Fields.Description}, "\n")
			request, err := http.NewRequest("POST", client.path, newRequestBody(strToAnalyze))
			if err != nil {
				return scores, err
			}
			resp, err := client.Do(request)
			if err != nil {
				return scores, err
			}
			defer resp.Body.Close()
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return scores, err
			}
			var jsonResp GrammarResponse
			err = json.Unmarshal(respBody, &jsonResp)
			if err != nil {
				return scores, err
			}
			scores = append(scores, float64(len(jsonResp.Matches)))
		}
		time.Sleep(1 * time.Minute)
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
	for i := 0; i < len(issues); i += gcpRateLimit {
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

// MultipleScores takes multiple issues and scorers and returns a map for each scorer to its corresponding scores.
func MultipleScores(issues []jira.Issue, scorers ...Scorer) (map[string][]float64, error) {
	scoreMap := make(map[string][]float64)
	var wg sync.WaitGroup
	var err error
	for _, scorer := range scorers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scores, e := scorer.Scores(issues...)
			if e != nil {
				err = e
			}
			scoreMap[scorer.Name()] = scores
		}()
		if err != nil {
			break
		}
	}
	wg.Wait()
	return scoreMap, err
}
