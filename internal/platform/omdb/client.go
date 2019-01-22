package omdb

import (
	"net/http"
	"time"
)

// Config contains the configuration items for an omdb client
type Config struct {
	BaseURL string `envconfig:"BASE_URL" required:"true"`
	APIKey  string `envconfig:"API_KEY" required:"true"`
	Timeout int    `envconfig:"TIMEOUT" required:"true"`
}

// Client contains the http client and url of the obdb client
type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// NewClient creates a new omdb client using the given configuration
func NewClient(config Config) *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * time.Duration(config.Timeout),
		},
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
	}
}
