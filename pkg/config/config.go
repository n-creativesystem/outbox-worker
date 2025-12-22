package config

import (
	"context"
	"database/sql"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Database  Database   `yaml:"database"`
	SSH       *SSH       `yaml:"ssh"`
	Publisher *Publisher `yaml:"publisher"`
	Logging   Logging    `yaml:"logging"`
	Tracking  Tracking   `yaml:"tracking"`
}

func (c *Config) Build() (string, error) {
	return c.Database.Build()
}

func (c *Config) Connect(ctx context.Context) (*sql.DB, error) {
	dsn, err := c.Build()
	if err != nil {
		return nil, err
	}
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	setDB(conn, c.Database)
	if err := conn.PingContext(ctx); err != nil {
		return nil, err
	}
	return conn, nil
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Database),
		validation.Field(&c.SSH),
		validation.Field(&c.Publisher, validation.NotNil),
		validation.Field(&c.Logging),
		validation.Field(&c.Tracking),
	)
}
