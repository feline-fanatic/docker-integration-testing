package omdb

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var config struct {
	OmdbConfig Config `envconfig:"OMDB" required:"true"`
}

func TestMain(m *testing.M) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := envconfig.Process("", &config); err != nil {
		logrus.WithError(err).Fatal("error while processing env vars")
	}

	os.Exit(m.Run())
}

func TestGetRatings(t *testing.T) {
	tests := []struct {
		title           string
		expectedRatings []Rating
	}{
		{
			"NineLives",
			[]Rating{
				{Source: "Internet Movie Database", Value: "5.3/10"},
				{Source: "Rotten Tomatoes", Value: "14%"},
				{Source: "Metacritic", Value: "11/100"},
			},
		},
		{
			"Aristocats",
			[]Rating{
				{Source: "Internet Movie Database", Value: "7.1/10"},
				{Source: "Rotten Tomatoes", Value: "68%"},
				{Source: "Metacritic", Value: "66/100"},
			},
		},
		{
			"Keanu",
			[]Rating{
				{Source: "Internet Movie Database", Value: "6.3/10"},
				{Source: "Rotten Tomatoes", Value: "77%"},
				{Source: "Metacritic", Value: "63/100"},
			},
		},
		{
			"Garfield",
			[]Rating{
				{Source: "Internet Movie Database", Value: "5.0/10"},
				{Source: "Rotten Tomatoes", Value: "15%"},
				{Source: "Metacritic", Value: "27/100"},
			},
		},
	}

	client := NewClient(config.OmdbConfig)

	for _, tt := range tests {
		ratings, err := client.GetRatings(tt.title)
		if err != nil {
			t.Fatalf("error while calling GetRatings, err: %v", err)
		}
		for i, r := range ratings {
			if r.Source != tt.expectedRatings[i].Source || r.Value != tt.expectedRatings[i].Value {
				t.Errorf("source or value doesn't match")
			}
		}
	}
}
