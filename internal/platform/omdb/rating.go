package omdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Ratings is the response structure for ratings from a request to the omdb api
type Ratings struct {
	Ratings []Rating `json:"Ratings"`
}

// Rating is the structure of a rating
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

// GetRatings gets the ratings for a movie
func (c *Client) GetRatings(movie string) ([]Rating, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/?apiKey=%s&t=%s", c.baseURL, c.apiKey, url.QueryEscape(movie)))
	if err != nil {
		return nil, errors.Wrap(err, "error while calling omdb api")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error occurred while calling omdb api: %s, status code: %d", http.StatusText(resp.StatusCode), resp.StatusCode)
	}

	var ratings Ratings
	if err := json.NewDecoder(resp.Body).Decode(&ratings); err != nil {
		return nil, errors.Wrap(err, "error while decoding response from omdb api")
	}

	return ratings.Ratings, nil
}
