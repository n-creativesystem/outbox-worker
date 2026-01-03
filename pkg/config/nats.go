package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/nats-io/nats.go"
)

type NatsTLS struct {
	ServerName string   `yaml:"serverName"`
	RootCas    []string `yaml:"rootCAs"`
	ClientCert string   `yaml:"clientCert"`
	ClientKey  string   `yaml:"clientKey"`
}

func (c *NatsTLS) Config() (*tls.Config, error) {
	certFile := c.ClientCert
	keyFile := c.ClientKey
	var (
		certs []tls.Certificate
		err   error
	)
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("error parsing X509 certificate/key pair: %v", err)
		}
		certs = append(certs, cert)
	}
	rootCAsCB := func() (*x509.CertPool, error) {
		pool := x509.NewCertPool()
		for _, f := range c.RootCas {
			rootPEM, err := os.ReadFile(f)
			if err != nil || rootPEM == nil {
				return nil, fmt.Errorf("nats: error loading or parsing rootCA file: %w", err)
			}
			ok := pool.AppendCertsFromPEM(rootPEM)
			if !ok {
				return nil, fmt.Errorf("nats: failed to parse root certificate from %q", f)
			}
		}
		return pool, nil
	}

	pool, err := rootCAsCB()
	if err != nil {
		return nil, fmt.Errorf("error parsing X509 root ca: %v", err)
	}

	tlsConfig := &tls.Config{
		ServerName:   c.ServerName,
		Certificates: certs,
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

type NatsUserCredentials struct {
	UserOrChainedFile string   `yaml:"userOrChainedFile"`
	SeedFiles         []string `yaml:"seedFiles"`
}

func (c NatsUserCredentials) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.UserOrChainedFile, validation.Required),
		validation.Field(&c.SeedFiles, validation.Length(1, 0)),
	)
}

type NatsUserInfo struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func (c NatsUserInfo) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.Password, validation.Required),
	)
}

type NATS struct {
	ClientName string `yaml:"clientName"`
	Server     string `yaml:"server"`

	MaxReconnects      int           `yaml:"maxReconnects" default:"5"`
	ReconnectWait      time.Duration `yaml:"reconnectWait" default:"2s"`
	ReconnectJitter    time.Duration `yaml:"reconnectJitter" default:"100ms"`
	ReconnectJitterTLS time.Duration `yaml:"reconnectJitterTLS" default:"1s"`
	Timeout            time.Duration `yaml:"timeout" default:"5s"`

	TLS             *NatsTLS             `yaml:"tls"`
	UserInfo        *NatsUserInfo        `yaml:"userInfo"`
	UserCredentials *NatsUserCredentials `yaml:"userCredentials"`
	Token           *string              `yaml:"token"`
	Domain          *string              `yaml:"domain"`
}

func (c NATS) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Server, validation.Required),
		validation.Field(&c.Domain, validation.NilOrNotEmpty),
		validation.Field(&c.TLS),
		validation.Field(&c.UserInfo),
		validation.Field(&c.UserCredentials),
		validation.Field(&c.Token, validation.NilOrNotEmpty),
	)
}

func (c *NATS) Options() ([]nats.Option, error) {
	opt := make([]nats.Option, 0, 20)
	opt = append(opt,
		nats.Name(c.ClientName),
		nats.MaxReconnects(c.MaxReconnects),
		nats.ReconnectWait(c.ReconnectWait),
		nats.ReconnectJitter(c.ReconnectJitter, c.ReconnectJitterTLS),
		nats.Timeout(c.Timeout),
	)
	if tls := c.TLS; tls != nil {
		if tlsConf, err := tls.Config(); err != nil {
			return nil, err
		} else {
			opt = append(opt, nats.Secure(tlsConf))
		}
	}
	if cred := c.UserCredentials; cred != nil {
		opt = append(opt, nats.UserCredentials(cred.UserOrChainedFile, cred.SeedFiles...))
	}
	if info := c.UserInfo; info != nil {
		opt = append(opt, nats.UserInfo(info.User, info.Password))
	}
	if token := c.Token; token != nil {
		opt = append(opt, nats.Token(*token))
	}
	return opt, nil
}
