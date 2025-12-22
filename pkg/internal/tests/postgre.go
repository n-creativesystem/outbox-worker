package tests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	container    *postgrescontainer.PostgresContainer
	adminDB      *sql.DB
	postgresConf *pgx.ConnConfig
	mu           sync.Mutex
}

func (c *PostgresContainer) Terminate(ctx context.Context) error {
	var rErr error
	if err := c.adminDB.Close(); err != nil {
		rErr = errors.Join(rErr, err)
	}
	if err := c.container.Terminate(ctx); err != nil {
		rErr = errors.Join(rErr, err)
	}
	return rErr
}

func (c *PostgresContainer) Config(ctx context.Context) (*pgx.ConnConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dbName := fmt.Sprintf("outbox_%s", strings.ReplaceAll(uuid.NewString(), "-", ""))
	cfg := *c.postgresConf
	cfg.Database = dbName
	dsn := pgxConfigToDSN(cfg.Copy(), dbName)
	cloneCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if _, err := c.adminDB.ExecContext(ctx, "CREATE DATABASE "+dbName); err != nil {
		return nil, fmt.Errorf("CREATE DATABASE: %w", err)
	}
	return cloneCfg, nil
}

func pgxConfigToDSN(conf *pgx.ConnConfig, dbName string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(conf.User, conf.Password),
		Host:   fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Path:   "/" + dbName, // 先頭にスラッシュが必要です
	}

	q := u.Query()

	for k, v := range conf.RuntimeParams {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func (c *PostgresContainer) Open(cfg *pgx.ConnConfig) (*sql.DB, error) {
	drv, err := sql.Open("pgx", cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %w", err)
	}
	return drv, nil
}

func (c *PostgresContainer) Client(t *testing.T, ctx context.Context) (*sql.DB, error) {
	cfg, err := c.Config(ctx)
	if err != nil {
		return nil, err
	}
	return c.Open(cfg)
}

func RunPostgreSQLContainer(ctx context.Context) (*PostgresContainer, error) {
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:16-alpine",
		postgrescontainer.WithDatabase("outbox"),
		postgrescontainer.WithUsername("admin"),
		postgrescontainer.WithPassword("pass1234"),
		testcontainers.WithAdditionalWaitStrategy(
			wait.ForLog("listening on IPv4 address \"0.0.0.0\", port 5432").
				AsRegexp().
				WithStartupTimeout(5*time.Minute).WithOccurrence(1),
		),
		testcontainers.WithLogger(log.Default()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	conn, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate postgres connection string: %w", err)
	}

	postgresCfg, err := pgx.ParseConfig(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to generate postgres connection string: %w", err)
	}

	adminDB, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %w", err)
	}

	if err := adminDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return &PostgresContainer{
		container:    postgresContainer,
		adminDB:      adminDB,
		postgresConf: postgresCfg,
		mu:           sync.Mutex{},
	}, nil
}
