package tests

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/stretchr/testify/require"
)

// MySQLProvider はMySQL用のテストDBプロバイダー
type MySQLProvider struct {
	container *MySQLContainer
}

// NewMySQLProvider は新しいMySQLProviderを作成する
func NewMySQLProvider(container *MySQLContainer) *MySQLProvider {
	return &MySQLProvider{
		container: container,
	}
}

func (p *MySQLProvider) Name() string {
	return "MySQL"
}

func (p *MySQLProvider) GetDialect() string {
	return "mysql"
}

func (p *MySQLProvider) CreateTestDB(t *testing.T) (*sql.DB, *config.OutboxPolling) {
	t.Helper()

	if p.container == nil {
		t.Fatal("MySQL container is not initialized")
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
	connStr := "mysql://" + cfg.FormatDSN()

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

func (p *MySQLProvider) CreateSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := t.Context()

	// outboxテーブルのスキーマを作成
	sqlScript, err := GetFileWithError("mysql_outbox_table.sql")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, string(sqlScript))
	require.NoError(t, err)
}

// ReplacePlaceholders はMySQLの場合、何も変換しない（'?'をそのまま使用）
func (p *MySQLProvider) ReplacePlaceholders(sql string) string {
	return sql
}

// InsertTestData はテストデータをMySQLに挿入する
func (p *MySQLProvider) InsertTestData(t *testing.T, db *sql.DB, rows [][]any) {
	t.Helper()

	if len(rows) == 0 {
		return
	}

	ctx := t.Context()

	// MySQLは複数行INSERTをサポートしているので、1回のクエリで挿入
	// VALUES (?, ?, ...), (?, ?, ...), ...
	placeholders := make([]string, len(rows))
	args := make([]any, 0, len(rows)*8)

	for i, row := range rows {
		p := make([]string, len(row))
		for j := range len(row) {
			p[j] = "?"
		}
		placeholders[i] = "(" + strings.Join(p, ",") + ")"
		args = append(args, row...)
	}

	query := `INSERT INTO outbox (aggregate_type, aggregate_id, event, payload, retry_at, retry_count, sent_at) VALUES ` +
		strings.Join(placeholders, ", ")

	_, err := db.ExecContext(ctx, query, args...)
	require.NoError(t, err)
}
