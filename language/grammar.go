package language

import (
	"encoding/json"
	"github.com/nclandrei/L5-Project/jira"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	clientRateLimit     = 20                                      // defines number of requests permitted per minute
	languageToolAPIPath = "https://languagetool.org/api/v2/check" // URL path to LanguageTool API
)

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
		rateLimit: clientRateLimit,
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

// Scores returns the grammar scores for all issues passed as arguments.
func (client *GrammarClient) Scores(issues ...jira.Issue) ([]int, error) {
	var scores []int
	for i := 0; i < len(issues); i += clientRateLimit {
		for _, issue := range issues[i:(i + clientRateLimit)] {
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
			scores = append(scores, len(jsonResp.Matches))
		}
		time.Sleep(1 * time.Minute)
	}
	return scores, nil
}
