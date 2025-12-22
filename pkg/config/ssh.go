package config

import (
	"fmt"
	"os"
	"path/filepath"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSH struct {
	PrivateKey        string   `yaml:"privateKey"`
	Host              string   `yaml:"host"`
	Port              int      `yaml:"port" default:"22"`
	Username          string   `yaml:"username"`
	HostKeyAlgorithms []string `yaml:"hostKeyAlgorithms" default:"[\"ssh-ed25519\"]"`
	KnownHosts        string   `yaml:"knownHosts"`
}

func (s *SSH) SetDefaults() {
	if s.KnownHosts == "" {
		s.KnownHosts = filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	}
}

func (conf *SSH) Conn() (*ssh.Client, error) {
	sshKey, err := os.ReadFile(conf.PrivateKey)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		return nil, err
	}
	hostKeyCallback, err := knownhosts.New(conf.KnownHosts)
	if err != nil {
		return nil, err
	}

	sshConf := &ssh.ClientConfig{
		User: conf.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   hostKeyCallback,
		HostKeyAlgorithms: conf.HostKeyAlgorithms,
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", conf.Host, conf.Port), sshConf)
}

func (conf SSH) Validate() error {
	return validation.ValidateStruct(&conf,
		validation.Field(&conf.PrivateKey, validation.Required),
		validation.Field(&conf.Host, validation.Required),
		validation.Field(&conf.Port, validation.Required),
	)
}
