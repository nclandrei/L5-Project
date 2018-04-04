package language

import (
	"encoding/json"
	"github.com/nclandrei/L5-Project/jira"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	clientRateLimit     = 20
	languageToolAPIPath = "https://languagetool.org/api/v2/check"
)

// Client defines the LanguageTool http client.
type Client struct {
	*http.Client
	rateLimit int
	path      string
}

// Response defines the response retrieved via LanguageTool API.
type Response struct {
	Matches []Match `json:"matches"`
}

// Match defines a match for an issue found in the parsed text.
type Match struct {
	Rule Rule `json:"rule"`
}

// Rule defines all the necessary info needed to understand a grammar error from LanguageTool.
type Rule struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	IssueType   string `json:"issueType"`
	Category    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}

// NewGrammarClient returns a new Grammar client.
func NewGrammarClient() *Client {
	return &Client{
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
func (client *Client) Scores(issues ...jira.Issue) ([]float64, error) {
	var scores []float64
	for _, issue := range issues {
		strToAnalyze := strings.Join([]string{issue.Fields.Summary, issue.Fields.Description}, "\n")
		request, err := http.NewRequest("POST", client.path, newRequestBody(strToAnalyze))
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var jsonResp Response
		err = json.Unmarshal(respBody, &jsonResp)
		if err != nil {
			return nil, err
		}
		log.Println(jsonResp)
	}
	return scores, nil
}
