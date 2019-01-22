package sftp

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var config struct {
	SftpConfig         Config `envconfig:"SFTP" required:"true"`
	TestPrivateKeyFile string `envconfig:"TEST_PRIVATE_KEY_FILE" required:"false"`
}

func TestMain(m *testing.M) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := envconfig.Process("", &config); err != nil {
		logrus.WithError(err).Fatal("error while processing env vars")
	}

	// get private test key
	testKey, err := ioutil.ReadFile(config.TestPrivateKeyFile)
	if err != nil {
		logrus.WithError(err).Fatal("error while reading test private key file")
	}
	config.SftpConfig.PrivateKey = string(testKey)

	os.Exit(m.Run())
}

func TestNewClient(t *testing.T) {
	client, err := NewClient(config.SftpConfig)
	if err != nil {
		t.Fatalf("error while creating sftp client %v", err)
	}
	client.Close()
}
