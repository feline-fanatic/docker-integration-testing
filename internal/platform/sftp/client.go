package sftp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Config contains the configuration items for an sftp client
type Config struct {
	Env        string `envconfig:"ENV" required:"false"` // only needed for integration tests
	Host       string `envconfig:"HOST" required:"true"`
	Port       string `envconfig:"PORT" required:"true"`
	User       string `envconfig:"USER" required:"true"`
	PrivateKey string `envconfig:"PRIVATE_KEY" required:"true"`
	Passphrase string `envconfig:"PASSPHRASE" required:"true"`
	HostKey    string `envconfig:"HOST_KEY" required:"true"`
	FilePath   string `envconfig:"FILE_PATH" required:"true"`
	Timeout    int    `envconfig:"TIMEOUT" required:"true"`
}

// NewClient creates a new sftp client using the given configuration
func NewClient(config Config) (*sftp.Client, error) {
	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(config.PrivateKey), []byte(config.Passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing private key")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: verify(config.HostKey, config.Env),
		Timeout:         time.Second * time.Duration(config.Timeout),
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.Host, config.Port), sshConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error while establishing ssh connection for sftp")
	}

	return sftp.NewClient(sshClient)
}

// verify verifies the host key of the ssh server
func verify(hostKey, env string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// for local tests
		if env == "local" && hostname == "sftp:22" {
			return nil
		}

		decoding, err := base64.StdEncoding.DecodeString(hostKey)
		if err != nil {
			return err
		}

		keyToCheckAgainst, err := ssh.ParsePublicKey([]byte(decoding))
		if err != nil {
			return err
		}

		if !bytes.Equal(keyToCheckAgainst.Marshal(), key.Marshal()) {
			return errors.New("ssh: host key mismatch")
		}

		return nil
	}
}
