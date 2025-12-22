package tests

import (
	"database/sql"
	"strconv"
	"strings"
	"testing"

	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/stretchr/testify/require"
)

// MySQLProvider はMySQL用のテストDBプロバイダー
type PostgresProvider struct {
	container *PostgresContainer
}

// NewMySQLProvider は新しいMySQLProviderを作成する
func NewPostgresProvider(container *PostgresContainer) *PostgresProvider {
	return &PostgresProvider{
		container: container,
	}
}

func (p *PostgresProvider) Name() string {
	return "Postgres"
}

func (p *PostgresProvider) GetDialect() string {
	return "pgx"
}

func (p *PostgresProvider) CreateTestDB(t *testing.T) (*sql.DB, *config.OutboxPolling) {
	t.Helper()

	if p.container == nil {
		t.Fatal("Postgres container is not initialized")
	}

	ctx := t.Context()

	// containerからデータベース設定を取得（各テストで一意のDB名が生成される）
	cfg, err := p.container.Config(ctx)
	require.NoError(t, err)

	// データベースを開く
	db, err := p.container.Open(cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	// 接続文字列を生成
	connStr := cfg.ConnString()

	// OutboxPolling設定を作成
	outboxCfg := &config.OutboxPolling{
		Config: config.Config{
			Database: config.Database{
				URI: connStr,
			},
		},
		OutboxConfig: &config.OutboxPollingConfig{
			TableName:      "outbox",
			ProducerName:   "test-producer",
			MaxRetryCount:  10,
			FindEventLimit: 100,
		},
	}

	return db, outboxCfg
}

func (p *PostgresProvider) CreateSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := t.Context()

	// outboxテーブルのスキーマを作成
	sqlScript, err := GetFileWithError("postgres_outbox_table.sql")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, string(sqlScript))
	require.NoError(t, err)
}

// ReplacePlaceholders はMySQLの場合、何も変換しない（'?'をそのまま使用）
func (p *PostgresProvider) ReplacePlaceholders(sql string) string {
	return sql
}

// InsertTestData はテストデータをMySQLに挿入する
func (p *PostgresProvider) InsertTestData(t *testing.T, db *sql.DB, rows [][]any) {
	t.Helper()

	if len(rows) == 0 {
		return
	}

	ctx := t.Context()

	// MySQLは複数行INSERTをサポートしているので、1回のクエリで挿入
	// VALUES (?, ?, ...), (?, ?, ...), ...
	placeholders := make([]string, len(rows))
	args := make([]any, 0, len(rows)*8)
	j := 1
	for i, row := range rows {
		p := make([]string, len(row))
		for i := range len(row) {
			p[i] = "$" + strconv.Itoa(j)
			j++
		}
		placeholders[i] = "(" + strings.Join(p, ",") + ")"
		args = append(args, row...)
	}

	query := `INSERT INTO outbox (aggregate_type, aggregate_id, event, payload, retry_at, retry_count, sent_at) VALUES ` +
		strings.Join(placeholders, ", ")

	_, err := db.ExecContext(ctx, query, args...)
	require.NoError(t, err)
}
