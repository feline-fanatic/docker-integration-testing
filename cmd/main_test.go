package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/feline-fanatic/docker-integration-testing/internal/platform/omdb"
	sftpClient "github.com/feline-fanatic/docker-integration-testing/internal/platform/sftp"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	awsresolver "stash.redventures.net/de/go-awsresolver"
)

var testConfig struct {
	SftpConfig         sftpClient.Config `envconfig:"SFTP" required:"true"`
	TestPrivateKeyFile string            `envconfig:"TEST_PRIVATE_KEY_FILE" required:"false"`
	OmdbConfig         omdb.Config       `envconfig:"OMDB" required:"true"`
	S3Region           string            `envconfig:"S3_REGION" required:"true"`
	S3Bucket           string            `envconfig:"S3_BUCKET" required:"true"`
}
var (
	uploader *s3manager.Uploader
	s3C      *s3.S3
	sftpC    *sftp.Client
	omdbC    *omdb.Client
)

func TestMain(m *testing.M) {
	var err error

	// get env vars
	if err = envconfig.Process("", &testConfig); err != nil {
		logrus.WithError(err).Fatal("error while configuring environment")
	}

	// get aws session
	awsSess, err := awsresolver.GetAWSSession(aws.Config{Region: &testConfig.S3Region})
	if err != nil {
		logrus.WithError(err).Fatal("error while getting aws session")
	}
	uploader = s3manager.NewUploader(awsSess)

	// set up s3 and create test bucket
	s3C = s3.New(awsSess)
	if _, err := s3C.CreateBucket(&s3.CreateBucketInput{Bucket: &testConfig.S3Bucket}); err != nil {
		logrus.WithError(err).Fatal("error while creating s3 bucket")
	}

	// get test private key
	testKey, err := ioutil.ReadFile(testConfig.TestPrivateKeyFile)
	if err != nil {
		logrus.WithError(err).Fatal("error while reading test private key file")
	}
	testConfig.SftpConfig.PrivateKey = string(testKey)

	// create sftp client
	sftpC, err = sftpClient.NewClient(testConfig.SftpConfig)
	if err != nil {
		logrus.WithError(err).Fatal("error whie creating sftp client")
	}

	// create omdb client
	omdbC = omdb.NewClient(testConfig.OmdbConfig)

	// cleanup
	defer sftpC.Close()

	os.Exit(m.Run())
}

func TestOpenSFTPFile(t *testing.T) {
	file := "/data/movie-list.json"
	f, err := openSFTPFile(file, sftpC)
	if err != nil {
		t.Errorf("cannot open file %s", file)
	}
	f.Close()
}

func TestGetRatingsForMovies(t *testing.T) {
	movies := []string{
		"NineLives",
		"Aristocats",
		"Keanu",
		"Garfield",
	}

	expectedData := RatingList{
		Movies: []MovieRating{
			{Name: "NineLives", Imdb: "5.3/10", RottenTomatoes: "14%", Metacritic: "11/100"},
			{Name: "Aristocats", Imdb: "7.1/10", RottenTomatoes: "68%", Metacritic: "66/100"},
			{Name: "Keanu", Imdb: "6.3/10", RottenTomatoes: "77%", Metacritic: "63/100"},
			{Name: "Garfield", Imdb: "5.0/10", RottenTomatoes: "15%", Metacritic: "27/100"},
		},
	}

	ratingList, err := getRatingsForMovies(movies, omdbC)
	if err != nil {
		t.Errorf("error while getting ratings for movies, err: %v", err)
	}

	for i, r := range ratingList.Movies {
		if r.Name == expectedData.Movies[i].Name && (r.Imdb != expectedData.Movies[i].Imdb ||
			r.RottenTomatoes != expectedData.Movies[i].RottenTomatoes || r.Metacritic != expectedData.Movies[i].Metacritic) {
			t.Error("doesn't match expected rating")
		}
	}
}

func TestUploadToS3(t *testing.T) {
	key := "omdb.ratings"
	data := RatingList{
		Movies: []MovieRating{
			{Name: "NineLives", Imdb: "5.3/10", RottenTomatoes: "14%", Metacritic: "11/100"},
			{Name: "Aristocats", Imdb: "7.1/10", RottenTomatoes: "68%", Metacritic: "66/100"},
			{Name: "Keanu", Imdb: "6.3/10", RottenTomatoes: "77%", Metacritic: "63/100"},
			{Name: "Garfield", Imdb: "5.0/10", RottenTomatoes: "15%", Metacritic: "27/100"},
		},
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		t.Fatalf("error while encoding record to json, err %v", err)
	}

	if err := uploadToS3(key, buf, uploader); err != nil {
		t.Fatalf("error while uploading to s3, err: %v", err)
	}

	output, err := s3C.GetObject(&s3.GetObjectInput{Bucket: &testConfig.S3Bucket, Key: &key})
	if err != nil {
		t.Errorf("error while validating that object exists in s3, err: %v", err)
	}

	var ratingList RatingList
	if err := json.NewDecoder(output.Body).Decode(&ratingList); err != nil {
		t.Errorf("error while decoding object, err: %v", err)
	}

	for i, r := range ratingList.Movies {
		if r.Name == data.Movies[i].Name && (r.Imdb != data.Movies[i].Imdb || r.RottenTomatoes != data.Movies[i].RottenTomatoes || r.Metacritic != data.Movies[i].Metacritic) {
			t.Error("object from s3 doesnt match")
		}
	}
}

func TestProcess(t *testing.T) {
	if err := process(sftpC, omdbC, uploader); err != nil {
		t.Errorf("error while processing")
	}
}
