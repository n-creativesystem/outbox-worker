package tests

import (
	"database/sql"
	"testing"

	"github.com/n-creativesystem/outbox-worker/pkg/config"
)

type DBProvider interface {
	Name() string
	CreateTestDB(t *testing.T) (*sql.DB, *config.OutboxPolling)
	GetDialect() string
	CreateSchema(t *testing.T, db *sql.DB)
	ReplacePlaceholders(sql string) string
	InsertTestData(t *testing.T, db *sql.DB, rows [][]any)
}
