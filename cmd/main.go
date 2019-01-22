package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/feline-fanatic/docker-integration-testing/internal/platform/omdb"
	sftpClient "github.com/feline-fanatic/docker-integration-testing/internal/platform/sftp"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	awsresolver "stash.redventures.net/de/go-awsresolver"
)

var config struct {
	SftpConfig sftpClient.Config `envconfig:"SFTP" required:"true"`
	OmdbConfig omdb.Config       `envconfig:"OMDB" required:"true"`
	S3Region   string            `envconfig:"S3_REGION" required:"true"`
	S3Bucket   string            `envconfig:"S3_BUCKET" required:"true"`
}

// MovieList contains a list of movies
type MovieList struct {
	Movies []string `json:"movies"`
}

// RatingList contains a list of MovieRatings
type RatingList struct {
	Movies []MovieRating `json:"movies"`
}

// MovieRating contains the name of a movie and it's ratings
type MovieRating struct {
	Name           string `json:"name"`
	Imdb           string `json:"imdb"`
	RottenTomatoes string `json:"rottenTomatoes"`
	Metacritic     string `json:"metacritic"`
}

func init() {
	// set up logger
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// get env vars
	if err := envconfig.Process("", &config); err != nil {
		logrus.WithError(err).Fatal("Unable to configure from environment")
	}
}

func main() {
	// get sftp client
	sftpC, err := sftpClient.NewClient(config.SftpConfig)
	if err != nil {
		logrus.WithError(err).Fatal("error while creating sftp client")
	}
	defer sftpC.Close()

	// get omdb client
	omdbC := omdb.NewClient(config.OmdbConfig)

	// get aws session and s3 uploader
	awsSess, err := awsresolver.GetAWSSession(aws.Config{Region: &config.S3Region})
	if err != nil {
		logrus.WithError(err).Fatal("error while creating aws session")
	}
	uploader := s3manager.NewUploader(awsSess)

	if err := process(sftpC, omdbC, uploader); err != nil {
		logrus.WithError(err).Error("cannot process file")
	}
}

// process processes a sftp file by opening it, parsing, uploading to s3
func process(sftpC *sftp.Client, omdbC *omdb.Client, uploader *s3manager.Uploader) error {
	logrus.Info("processing file")

	// open file
	logrus.WithField("sftp path", config.SftpConfig.FilePath).Info("opening file")
	file, err := openSFTPFile(config.SftpConfig.FilePath, sftpC)
	if err != nil {
		return errors.Wrap(err, "error while opening file")
	}
	defer file.Close()

	// parse records in file
	logrus.Info("parsing file")
	movies, err := parseFile(file)
	if err != nil {
		return errors.Wrap(err, "error while parsing file")
	}

	// get ratings for all movies in the list
	ratingList, err := getRatingsForMovies(movies, omdbC)
	if err != nil {
		return errors.Wrap(err, "error while getting ratings for list of movies")
	}

	fmt.Println(ratingList)

	// encode ratingList
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(ratingList); err != nil {
		return errors.Wrap(err, "error while encoding record to json")
	}

	// upload to s3
	s3Key := "omdb.ratings"
	logrus.WithFields(logrus.Fields{"key": s3Key}).Info("uploading to s3")
	if err := uploadToS3(s3Key, buf, uploader); err != nil {
		return errors.Wrap(err, "error while uploading to s3")
	}

	return nil
}

// openSFTPFile opens a file from sftp. caller is responsible for closing file
func openSFTPFile(filePath string, sftpC *sftp.Client) (*sftp.File, error) {
	file, err := sftpC.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	return file, nil
}

// parseFile parses a file containing a list of movies
func parseFile(file io.Reader) ([]string, error) {
	var movieList MovieList
	if err := json.NewDecoder(file).Decode(&movieList); err != nil {
		return nil, errors.Wrap(err, "error while decoding file")
	}
	return movieList.Movies, nil
}

// getRatingsForMovies gets all of the ratins for a list of movies
func getRatingsForMovies(movies []string, omdbC *omdb.Client) (*RatingList, error) {
	movieRatings := make([]MovieRating, len(movies))
	for i, m := range movies {
		ratings, err := omdbC.GetRatings(m)
		if err != nil {
			return nil, errors.Wrapf(err, "error while getting ratings for movie %s", m)
		}

		var imdb, rottenTomatoes, metacritic string
		for _, r := range ratings {
			switch r.Source {
			case "Internet Movie Database":
				imdb = r.Value
			case "Rotten Tomatoes":
				rottenTomatoes = r.Value
			case "Metacritic":
				metacritic = r.Value
			}
		}

		movieRatings[i] = MovieRating{
			Name:           m,
			Imdb:           imdb,
			RottenTomatoes: rottenTomatoes,
			Metacritic:     metacritic,
		}
	}

	return &RatingList{
		Movies: movieRatings,
	}, nil
}

// uploadToS3 uploads file to the s3 bucket specified in the config
func uploadToS3(key string, data io.Reader, uploader *s3manager.Uploader) error {
	if _, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &config.S3Bucket,
		Key:    &key,
		Body:   data,
	}); err != nil {
		return errors.Wrap(err, "error while uploading to s3")
	}
	return nil
}
