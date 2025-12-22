package tests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	mysqlcontainer "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MySQLContainer struct {
	container *mysqlcontainer.MySQLContainer
	adminDB   *sql.DB
	mysqlCfg  *mysql.Config
	mu        sync.Mutex
}

func (c *MySQLContainer) Terminate(ctx context.Context) error {
	var rErr error
	if err := c.adminDB.Close(); err != nil {
		rErr = errors.Join(rErr, err)
	}
	if err := c.container.Terminate(ctx); err != nil {
		rErr = errors.Join(rErr, err)
	}
	return rErr
}

func (c *MySQLContainer) Config(ctx context.Context) (*mysql.Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dbName := fmt.Sprintf("outbox_%s", strings.ReplaceAll(uuid.NewString(), "-", ""))
	mysqlCfg := c.mysqlCfg.Clone()
	mysqlCfg.DBName = dbName

	if _, err := c.adminDB.ExecContext(ctx, "CREATE DATABASE `"+dbName+"`"); err != nil {
		return nil, fmt.Errorf("CREATE DATABASE: %w", err)
	}
	return mysqlCfg, nil
}

func (c *MySQLContainer) Open(cfg *mysql.Config) (*sql.DB, error) {
	conn := cfg.FormatDSN()
	drv, err := sql.Open(dialect.MySQL, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %w", err)
	}
	return drv, nil
}

func (c *MySQLContainer) Client(t *testing.T, ctx context.Context) (*sql.DB, error) {
	cfg, err := c.Config(ctx)
	if err != nil {
		return nil, err
	}
	return c.Open(cfg)
}

func RunMySQLContainer(ctx context.Context) (*MySQLContainer, error) {
	initGrantScript, err := getGrantScript()
	if err != nil {
		return nil, fmt.Errorf("failed get grant script")
	}

	mysqlContainer, err := mysqlcontainer.Run(ctx,
		"mysql:8.4",
		mysqlcontainer.WithDatabase("outbox"),
		mysqlcontainer.WithUsername("admin"),
		mysqlcontainer.WithPassword("pass1234"),
		mysqlcontainer.WithScripts(initGrantScript),
		testcontainers.WithAdditionalWaitStrategy(
			wait.ForLog("ready for connections\\..*port: 3306").
				AsRegexp().
				WithStartupTimeout(5*time.Minute).WithOccurrence(1),
		),
		testcontainers.WithLogger(log.Default()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start mysql container: %w", err)
	}

	conn, err := mysqlContainer.ConnectionString(
		ctx,
		[]string{
			"parseTime=true",
			"loc=UTC",
			"collation=utf8mb4_bin",
			"charset=utf8mb4",
			"sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'",
		}...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mysql connection string: %w", err)
	}

	mysqlCfg, err := mysql.ParseDSN(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mysql connection string: %w", err)
	}

	adminDB, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %w", err)
	}

	if err := adminDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return &MySQLContainer{
		container: mysqlContainer,
		adminDB:   adminDB,
		mysqlCfg:  mysqlCfg,
	}, nil
}
