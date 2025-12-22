package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	dialect           string
	scheme            string
	URI               string `yaml:"uri"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	TLS               *TLS   `yaml:"tls"`
	SSH               *SSH   `yaml:"ssh"`
	MaxOpenConn       int    `yaml:"maxOpenConn" default:"10"`
	MaxLifeTimeSecond int    `yaml:"maxLifeTimeSecond" default:"300"`
	MaxIdleConn       int    `yaml:"maxIdleConn" default:"1"`
	MaxIdleSecond     int    `yaml:"maxIdleSecond" default:"0"`
}

var (
	_ validation.Validatable = (*Database)(nil)
)

func (d *Database) Build() (string, error) {
	schemas := strings.SplitN(d.URI, "://", 2)
	scheme := schemas[0]
	switch scheme {
	case "mysql":
		d.URI, _ = strings.CutPrefix(d.URI, "mysql://")
		d.dialect = scheme
		d.scheme = scheme
		return d.mysqlBuild()
	case "postgres":
		d.dialect = "pgx"
		return d.postgresBuild()
	default:
		return "", fmt.Errorf("unsupported scheme: %s", scheme)
	}
}

func (d *Database) Dialect() string {
	return d.dialect
}

func (d *Database) mysqlBuild() (string, error) {
	cfg, err := mysql.ParseDSN(d.URI)
	if err != nil {
		return "", err
	}
	if d.Username != "" {
		cfg.User = d.Username
	}
	if d.Password != "" {
		cfg.Passwd = d.Password
	}
	if d.SSH != nil {
		var dialFunc mysql.DialContextFunc
		sshClient, err := d.SSH.Conn()
		if err != nil {
			return "", err
		}
		mysqlNet := "mysql+tcp"
		dialFunc = func(ctx context.Context, addr string) (net.Conn, error) {
			return sshClient.Dial("tcp", addr)
		}
		mysql.RegisterDialContext(mysqlNet, dialFunc)
	}
	if d.TLS != nil {
		tlsConfig, err := d.TLS.Config()
		if err != nil {
			return "", err
		}
		if err := mysql.RegisterTLSConfig("custom", tlsConfig); err != nil {
			return "", err
		}
		cfg.TLSConfig = "custom"
	}
	return cfg.FormatDSN(), nil
}

func (d *Database) postgresBuild() (string, error) {
	p, err := pgxpool.ParseConfig(d.URI)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to parse postgres connection: %v", err))
		return d.URI, nil
	}
	return p.ConnString(), nil
}

func (d Database) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.URI, validation.Required, is.URL),
		validation.Field(&d.Username, validation.Required),
		validation.Field(&d.Password, validation.Required),
	)
}

type TLS struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	SeverName          string `yaml:"serverName"`
	CAFile             string `yaml:"caFile"`
}

func (t *TLS) Config() (*tls.Config, error) {
	cfg := &tls.Config{
		InsecureSkipVerify: t.InsecureSkipVerify, // nolint: gosec
		ServerName:         t.SeverName,
	}
	if t.CAFile != "" {
		caCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(t.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tls: %w", err)
		}
		if ok := caCertPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to append PEM")
		}
		cfg.RootCAs = caCertPool
	}
	return cfg, nil
}

func setDB(db *sql.DB, cfg Database) {
	if cfg.MaxIdleSecond > 0 {
		db.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleSecond) * time.Second)
	}
	if cfg.MaxLifeTimeSecond > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTimeSecond) * time.Second)
	}
	if cfg.MaxIdleConn > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConn)
	}
	if cfg.MaxOpenConn > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConn)
	}
}
