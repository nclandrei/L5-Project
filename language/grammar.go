package language

import (
	"net/http"
)

// Client defines the LanguageTool http client.
type Client struct {
	*http.Client
}
